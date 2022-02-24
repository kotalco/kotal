package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	bitcoinv1alpha1 "github.com/kotalco/kotal/apis/bitcoin/v1alpha1"
	chainlinkv1alpha1 "github.com/kotalco/kotal/apis/chainlink/v1alpha1"
	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	nearv1alpha1 "github.com/kotalco/kotal/apis/near/v1alpha1"
	polkadotv1alpha1 "github.com/kotalco/kotal/apis/polkadot/v1alpha1"
	bitcoincontroller "github.com/kotalco/kotal/controllers/bitcoin"
	chainlinkcontroller "github.com/kotalco/kotal/controllers/chainlink"
	ethereumcontroller "github.com/kotalco/kotal/controllers/ethereum"
	ethereum2controller "github.com/kotalco/kotal/controllers/ethereum2"
	filecoincontroller "github.com/kotalco/kotal/controllers/filecoin"
	ipfscontroller "github.com/kotalco/kotal/controllers/ipfs"
	nearcontroller "github.com/kotalco/kotal/controllers/near"
	polkadotcontroller "github.com/kotalco/kotal/controllers/polkadot"
	// +kubebuilder:scaffold:imports
)

var (
	scheme         = runtime.NewScheme()
	setupLog       = ctrl.Log.WithName("setup")
	enableWebhooks = os.Getenv("ENABLE_WEBHOOKS") != "false"
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = ethereumv1alpha1.AddToScheme(scheme)
	_ = ethereum2v1alpha1.AddToScheme(scheme)
	_ = ipfsv1alpha1.AddToScheme(scheme)
	_ = filecoinv1alpha1.AddToScheme(scheme)
	_ = polkadotv1alpha1.AddToScheme(scheme)
	_ = chainlinkv1alpha1.AddToScheme(scheme)
	_ = nearv1alpha1.AddToScheme(scheme)
	_ = bitcoinv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "2b1fce2f.kotal.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&filecoincontroller.NodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Node")
		os.Exit(1)
	}
	if enableWebhooks {
		if err = (&filecoinv1alpha1.Node{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Node")
			os.Exit(1)
		}
	}

	if err = (&ethereumcontroller.NodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Node")
		os.Exit(1)
	}
	if enableWebhooks {
		if err = (&ethereumv1alpha1.Node{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Node")
			os.Exit(1)
		}
	}

	if err = (&ethereum2controller.BeaconNodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BeaconNode")
		os.Exit(1)
	}
	if enableWebhooks {
		if err = (&ethereum2v1alpha1.BeaconNode{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Node")
			os.Exit(1)
		}
	}

	if err = (&ethereum2controller.ValidatorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Validator")
		os.Exit(1)
	}
	if enableWebhooks {
		if err = (&ethereum2v1alpha1.Validator{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Validator")
			os.Exit(1)
		}
	}

	if err = (&ipfscontroller.PeerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Peer")
		os.Exit(1)
	}
	if enableWebhooks {
		if err = (&ipfsv1alpha1.Peer{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Peer")
			os.Exit(1)
		}
	}

	if err = (&ipfscontroller.ClusterPeerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterPeer")
		os.Exit(1)
	}
	if enableWebhooks {
		if err = (&ipfsv1alpha1.ClusterPeer{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "ClusterPeer")
			os.Exit(1)
		}
	}

	if err = (&polkadotcontroller.NodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Node")
		os.Exit(1)
	}
	if enableWebhooks {
		if err = (&polkadotv1alpha1.Node{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Node")
			os.Exit(1)
		}
	}

	if err = (&chainlinkcontroller.NodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Node")
		os.Exit(1)
	}
	if enableWebhooks {
		if err = (&chainlinkv1alpha1.Node{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Node")
			os.Exit(1)
		}
	}

	if err = (&nearcontroller.NodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Node")
		os.Exit(1)
	}
	if enableWebhooks {
		if err = (&nearv1alpha1.Node{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Node")
			os.Exit(1)
		}
	}

	if err = (&bitcoincontroller.NodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Node")
		os.Exit(1)
	}
	if enableWebhooks {
		if err = (&bitcoinv1alpha1.Node{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Node")
			os.Exit(1)
		}
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
