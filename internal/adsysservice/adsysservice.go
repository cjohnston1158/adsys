// Package adsysservice is the implementation of all GRPC calls endpoints.
package adsysservice

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/sirupsen/logrus"
	"github.com/ubuntu/adsys"
	"github.com/ubuntu/adsys/internal/ad"
	"github.com/ubuntu/adsys/internal/ad/backends"
	"github.com/ubuntu/adsys/internal/ad/backends/sss"
	"github.com/ubuntu/adsys/internal/ad/backends/winbind"
	"github.com/ubuntu/adsys/internal/authorizer"
	"github.com/ubuntu/adsys/internal/consts"
	"github.com/ubuntu/adsys/internal/daemon"
	"github.com/ubuntu/adsys/internal/decorate"
	"github.com/ubuntu/adsys/internal/grpc/connectionnotify"
	"github.com/ubuntu/adsys/internal/grpc/interceptorschain"
	"github.com/ubuntu/adsys/internal/grpc/logconnections"
	log "github.com/ubuntu/adsys/internal/grpc/logstreamer"
	"github.com/ubuntu/adsys/internal/i18n"
	"github.com/ubuntu/adsys/internal/policies"
	"google.golang.org/grpc"
)

// Service is used to implement adsys.ServiceServer.
type Service struct {
	adsys.UnimplementedServiceServer
	logger *logrus.Logger

	adc           *ad.AD
	policyManager *policies.Manager

	authorizer authorizerer

	state          state
	initSystemTime *time.Time

	bus    *dbus.Conn
	daemon *daemon.Daemon
}

type state struct {
	cacheDir     string
	runDir       string
	dconfDir     string
	sudoersDir   string
	policyKitDir string
	apparmorDir  string
}

type options struct {
	cacheDir      string
	runDir        string
	dconfDir      string
	sudoersDir    string
	policyKitDir  string
	apparmorDir   string
	apparmorFsDir string
	adBackend     string
	sssConfig     sss.Config
	winbindConfig winbind.Config
	authorizer    authorizerer
}
type option func(*options) error

type authorizerer interface {
	IsAllowedFromContext(context.Context, authorizer.Action) error
}

// WithCacheDir specifies a personalized daemon cache directory.
func WithCacheDir(p string) func(o *options) error {
	return func(o *options) error {
		o.cacheDir = p
		return nil
	}
}

// WithRunDir specifies a personalized /run.
func WithRunDir(p string) func(o *options) error {
	return func(o *options) error {
		o.runDir = p
		return nil
	}
}

// WithDconfDir specifies a personalized /etc/dconf.
func WithDconfDir(p string) func(o *options) error {
	return func(o *options) error {
		o.dconfDir = p
		return nil
	}
}

// WithSudoersDir specifies a personalized sudoers directory.
func WithSudoersDir(p string) func(o *options) error {
	return func(o *options) error {
		o.sudoersDir = p
		return nil
	}
}

// WithPolicyKitDir specifies a personalized policykit directory.
func WithPolicyKitDir(p string) func(o *options) error {
	return func(o *options) error {
		o.policyKitDir = p
		return nil
	}
}

// WithApparmorDir specifies a personalized apparmor directory.
func WithApparmorDir(p string) func(o *options) error {
	return func(o *options) error {
		o.apparmorDir = p
		return nil
	}
}

// WithApparmorFsDir specifies a personalized directory for the apparmor
// security filesystem.
func WithApparmorFsDir(p string) func(o *options) error {
	return func(o *options) error {
		o.apparmorFsDir = p
		return nil
	}
}

// WithADBackend specifies our specific backend to select.
func WithADBackend(backend string) func(o *options) error {
	return func(o *options) error {
		o.adBackend = backend
		return nil
	}
}

// WithSSSConfig specifies our specific sss options to override.
func WithSSSConfig(c sss.Config) func(o *options) error {
	return func(o *options) error {
		o.sssConfig = c
		return nil
	}
}

// WithWinbindConfig specifies our specific winbind options to override.
func WithWinbindConfig(c winbind.Config) func(o *options) error {
	return func(o *options) error {
		o.winbindConfig = c
		return nil
	}
}

// New returns a new instance of an AD service.
// If url or domain is empty, we load the missing parameters from sssd.conf, taking first
// domain in the list if not provided.
func New(ctx context.Context, opts ...option) (s *Service, err error) {
	defer decorate.OnError(&err, i18n.G("couldn't create adsys service"))

	// defaults
	args := options{}
	// applied options
	for _, o := range opts {
		if err := o(&args); err != nil {
			return nil, err
		}
	}

	// Create run and cache base directories
	runDir := args.runDir
	if runDir == "" {
		runDir = consts.DefaultRunDir
	}
	// #nosec G301 - we need to ensure users have access directly to their own scripts
	if err := os.MkdirAll(runDir, 0755); err != nil {
		return nil, err
	}

	cacheDir := args.cacheDir
	if cacheDir == "" {
		cacheDir = consts.DefaultCacheDir
	}
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, err
	}

	// Don’t call dbus.SystemBus which caches globally system dbus (issues in tests)
	bus, err := dbus.SystemBusPrivate()
	if err != nil {
		return nil, err
	}
	if err = bus.Auth(nil); err != nil {
		_ = bus.Close()
		return nil, err
	}
	if err = bus.Hello(); err != nil {
		_ = bus.Close()
		return nil, err
	}

	var adOptions []ad.Option
	if args.cacheDir != "" {
		adOptions = append(adOptions, ad.WithCacheDir(args.cacheDir))
	}
	if args.runDir != "" {
		adOptions = append(adOptions, ad.WithRunDir(args.runDir))
	}

	// AD Backend selection
	var adBackend backends.Backend
	switch args.adBackend {
	default:
		log.Warningf(ctx, "Unknown configured backend %q. Defaulting to sssd.", args.adBackend)
		fallthrough
	case "":
		fallthrough
	case "sssd":
		adBackend, err = sss.New(ctx, args.sssConfig, bus)
	case "winbind":
		adBackend, err = winbind.New(ctx, args.winbindConfig)
	}
	if err != nil {
		return nil, fmt.Errorf(i18n.G("could not initialize AD backend: %v"), err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	// For machines where /proc/sys/kernel/hostname returns FQDN, cut it.
	hostname, _, _ = strings.Cut(hostname, ".")

	adc, err := ad.New(ctx, bus, adBackend, hostname, adOptions...)
	if err != nil {
		return nil, err
	}

	if args.authorizer == nil {
		args.authorizer, err = authorizer.New(bus)
		if err != nil {
			_ = bus.Close()
			return nil, err
		}
	}

	var policyOptions []policies.Option
	if args.cacheDir != "" {
		policyOptions = append(policyOptions, policies.WithCacheDir(args.cacheDir))
	}
	if args.dconfDir != "" {
		policyOptions = append(policyOptions, policies.WithDconfDir(args.dconfDir))
	}
	if args.sudoersDir != "" {
		policyOptions = append(policyOptions, policies.WithSudoersDir(args.sudoersDir))
	}
	if args.policyKitDir != "" {
		policyOptions = append(policyOptions, policies.WithPolicyKitDir(args.policyKitDir))
	}
	if args.runDir != "" {
		policyOptions = append(policyOptions, policies.WithRunDir(args.runDir))
	}
	if args.apparmorDir != "" {
		policyOptions = append(policyOptions, policies.WithApparmorDir(args.apparmorDir))
	}
	if args.apparmorFsDir != "" {
		policyOptions = append(policyOptions, policies.WithApparmorFsDir(args.apparmorFsDir))
	}
	m, err := policies.NewManager(bus, hostname, policyOptions...)
	if err != nil {
		return nil, err
	}

	// Init system reference time
	initSysTime := initSystemTime(bus)

	return &Service{
		adc:           adc,
		policyManager: m,
		authorizer:    args.authorizer,
		state: state{
			cacheDir:     args.cacheDir,
			dconfDir:     args.dconfDir,
			sudoersDir:   args.sudoersDir,
			policyKitDir: args.policyKitDir,
			runDir:       args.runDir,
			apparmorDir:  args.apparmorDir,
		},
		initSystemTime: initSysTime,
		bus:            bus,
	}, nil
}

// RegisterGRPCServer registers our service with the new interceptor chains.
// It will notify the daemon of any new connection.
func (s *Service) RegisterGRPCServer(d *daemon.Daemon) *grpc.Server {
	s.logger = logrus.StandardLogger()
	srv := grpc.NewServer(grpc.StreamInterceptor(
		interceptorschain.StreamServer(
			log.StreamServerInterceptor(s.logger),
			connectionnotify.StreamServerInterceptor(d),
			logconnections.StreamServerInterceptor(),
		)), authorizer.WithUnixPeerCreds())
	adsys.RegisterServiceServer(srv, s)
	s.daemon = d
	return srv
}

// Quit cleans every ressources than the service was using.
func (s *Service) Quit(ctx context.Context) {
	if err := s.bus.Close(); err != nil {
		log.Warningf(ctx, i18n.G("Can't disconnect system dbus: %v"), err)
	}
}

// initSystemTime returns systemd generator init system time.
func initSystemTime(bus *dbus.Conn) *time.Time {
	systemd := bus.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")
	val, err := systemd.GetProperty("org.freedesktop.systemd1.Manager.GeneratorsStartTimestamp")
	if err != nil {
		log.Warningf(context.Background(), "could not get system startup time? Can’t list next refresh: %v", err)
		return nil
	}
	start, ok := val.Value().(uint64)
	if !ok {
		log.Warningf(context.Background(), "invalid next system startup time: %v", val.Value())
		return nil
	}

	initSystemTime := time.Unix(int64(start)/1000000, 0)
	return &initSystemTime
}
