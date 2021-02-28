package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// ValidatorReconciler reconciles a Validator object
type ValidatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=validators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ethereum2.kotal.io,resources=validators/status,verbs=get;update;patch

// Reconcile reconciles Ethereum 2.0 validator client node
func (r *ValidatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	var validator ethereum2v1alpha1.Validator

	if err = r.Client.Get(ctx, req.NamespacedName, &validator); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	r.updateLabels(&validator)

	if err = r.reconcileValidatorDataPVC(&validator); err != nil {
		return
	}

	if err = r.reconcileValidatorConfigmap(&validator); err != nil {
		return
	}

	if err = r.reconcileValidatorStatefulset(&validator); err != nil {
		return
	}

	return
}

// updateLabels adds missing labels to the validator
func (r *ValidatorReconciler) updateLabels(validator *ethereum2v1alpha1.Validator) {

	if validator.Labels == nil {
		validator.Labels = map[string]string{}
	}

	validator.Labels["name"] = "node"
	validator.Labels["protocol"] = "ethereum2"
	validator.Labels["instance"] = validator.Name
}

// specValidatorConfigmap updates validator configmap spec
func (r *ValidatorReconciler) specValidatorConfigmap(validator *ethereum2v1alpha1.Validator, configmap *corev1.ConfigMap, definitions string) {

	configmap.ObjectMeta.Labels = validator.Labels

	if configmap.Data == nil {
		configmap.Data = map[string]string{}
	}

	configmap.Data["validator_definitions.yml"] = definitions
}

// reconcileValidatorConfigmap creates config map if it doesn't exist or update it
func (r *ValidatorReconciler) reconcileValidatorConfigmap(validator *ethereum2v1alpha1.Validator) error {

	// configmap is required for lighthouse validator definitions
	if validator.Spec.Client != ethereum2v1alpha1.LighthouseClient {
		return nil
	}

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      validator.Name,
			Namespace: validator.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, configmap, func() error {
		if err := ctrl.SetControllerReference(validator, configmap, r.Scheme); err != nil {
			r.Log.Error(err, "Unable to set controller reference on configmap")
			return err
		}

		definitions, err := CreateValidatorDefinitions(validator)
		if err != nil {
			return err
		}

		r.specValidatorConfigmap(validator, configmap, definitions)

		return nil
	})

	return err
}

// reconcileValidatorDataPVC reconciles node data persistent volume claim
func (r *ValidatorReconciler) reconcileValidatorDataPVC(validator *ethereum2v1alpha1.Validator) error {
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      validator.Name,
			Namespace: validator.Namespace,
		},
	}

	_, err := ctrl.CreateOrUpdate(context.Background(), r.Client, &pvc, func() error {
		if err := ctrl.SetControllerReference(validator, &pvc, r.Scheme); err != nil {
			return err
		}

		r.specValidatorDataPVC(validator, &pvc)

		return nil
	})

	return err
}

// specValidatorDataPVC updates node data PVC spec
func (r *ValidatorReconciler) specValidatorDataPVC(validator *ethereum2v1alpha1.Validator, pvc *corev1.PersistentVolumeClaim) {

	request := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(validator.Spec.Resources.Storage),
	}

	// spec is immutable after creation except resources.requests for bound claims
	if !pvc.CreationTimestamp.IsZero() {
		pvc.Spec.Resources.Requests = request
		return
	}

	pvc.Labels = validator.GetLabels()

	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		Resources: corev1.ResourceRequirements{
			Requests: request,
		},
		StorageClassName: validator.Spec.Resources.StorageClass,
	}
}

// createValidatorVolumes creates validator volumes
func (r *ValidatorReconciler) createValidatorVolumes(validator *ethereum2v1alpha1.Validator) (volumes []corev1.Volume) {

	var volumeProjections []corev1.VolumeProjection

	dataVolume := corev1.Volume{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: validator.Name,
			},
		},
	}
	volumes = append(volumes, dataVolume)

	// validator key/secret volumes
	for i, keystore := range validator.Spec.Keystores {

		var keystorePath string

		// Nimbus: looking for keystore.json files
		// Prysm: looking for keystore-{any suffix}.json files
		// lighthouse:
		//	- in auto discover mode: looking for voting-keystore.json files
		//	- in validator_defintions.yml: any file name or directory structure can be used
		// teku: indifferernt to file names or directory structure
		if validator.Spec.Client == ethereum2v1alpha1.NimbusClient {
			keystorePath = "keystore.json"
		} else {
			keystorePath = fmt.Sprintf("keystore-%d.json", i)
		}

		keystoreVolume := corev1.Volume{
			Name: keystore.SecretName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: keystore.SecretName,
					Items: []corev1.KeyToPath{
						{
							Key:  "keystore",
							Path: keystorePath,
						},
					},
				},
			},
		}

		// aggregate volume projections for nimbus client
		if validator.Spec.Client == ethereum2v1alpha1.NimbusClient {
			volumeProjections = append(volumeProjections, corev1.VolumeProjection{
				Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: keystore.SecretName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "password",
							Path: keystore.SecretName,
						},
					},
				},
			})
		} else {
			// update keystore volume with password for other clients
			keystoreVolume.VolumeSource.Secret.Items = append(keystoreVolume.VolumeSource.Secret.Items, corev1.KeyToPath{
				Key:  "password",
				Path: "password.txt",
			})
		}

		volumes = append(volumes, keystoreVolume)

	}

	if validator.Spec.Client == ethereum2v1alpha1.NimbusClient {
		// Nimbus checks if secrets dir has 0600 and tries to fix it if not
		mode := int32(0600)
		validatorSecretsVolume := corev1.Volume{
			Name: "validator-secrets",
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources:     volumeProjections,
					DefaultMode: &mode,
				},
			},
		}
		volumes = append(volumes, validatorSecretsVolume)
	}

	// lighthouse validator_definitions.yml
	if validator.Spec.Client == ethereum2v1alpha1.LighthouseClient {
		validatorDefinitionsVolume := corev1.Volume{
			// TODO: prepend validator name to avoid collision
			Name: "validator-definitions",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: validator.Name,
					},
				},
			},
		}
		volumes = append(volumes, validatorDefinitionsVolume)

	}

	// prysm wallet password volume
	if validator.Spec.Client == ethereum2v1alpha1.PrysmClient {
		walletPasswordVolume := corev1.Volume{
			Name: validator.Spec.WalletPasswordSecret,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: validator.Spec.WalletPasswordSecret,
					Items: []corev1.KeyToPath{
						{
							Key:  "password",
							Path: "prysm-wallet-password.txt",
						},
					},
				},
			},
		}
		volumes = append(volumes, walletPasswordVolume)
	}

	return
}

// createValidatorVolumeMounts creates validator volume mounts
func (r *ValidatorReconciler) createValidatorVolumeMounts(validator *ethereum2v1alpha1.Validator) (mounts []corev1.VolumeMount) {
	dataMount := corev1.VolumeMount{
		Name:      "data",
		MountPath: PathBlockchainData,
	}
	mounts = append(mounts, dataMount)

	for _, keystore := range validator.Spec.Keystores {

		keystoreMount := corev1.VolumeMount{
			Name:      keystore.SecretName,
			MountPath: fmt.Sprintf("%s/validator-keys/%s", PathSecrets, keystore.SecretName),
		}
		mounts = append(mounts, keystoreMount)
	}

	// Lighthouse validator_definitions.yml config
	if validator.Spec.Client == ethereum2v1alpha1.LighthouseClient {
		validatorDefinitionsMount := corev1.VolumeMount{
			Name:      "validator-definitions",
			MountPath: PathConfig,
		}
		mounts = append(mounts, validatorDefinitionsMount)
	}

	// prysm wallet password
	if validator.Spec.Client == ethereum2v1alpha1.PrysmClient {
		walletPasswordMount := corev1.VolumeMount{
			Name:      validator.Spec.WalletPasswordSecret,
			ReadOnly:  true,
			MountPath: fmt.Sprintf("%s/prysm-wallet", PathSecrets),
		}
		mounts = append(mounts, walletPasswordMount)
	}

	// nimbus client
	if validator.Spec.Client == ethereum2v1alpha1.NimbusClient {
		ValidatorSecretsMount := corev1.VolumeMount{
			Name:      "validator-secrets",
			ReadOnly:  true,
			MountPath: fmt.Sprintf("%s/validator-secrets", PathSecrets),
		}
		mounts = append(mounts, ValidatorSecretsMount)
	}

	return
}

// specValidatorStatefulset updates node statefulset spec
func (r *ValidatorReconciler) specValidatorStatefulset(validator *ethereum2v1alpha1.Validator, sts *appsv1.StatefulSet, img string, command, args []string) {

	sts.Labels = validator.GetLabels()

	initContainers := []corev1.Container{}
	mounts := r.createValidatorVolumeMounts(validator)

	// import validator keys into Prysm client
	if validator.Spec.Client == ethereum2v1alpha1.PrysmClient {
		for i, keystore := range validator.Spec.Keystores {
			keyDir := fmt.Sprintf("%s/validator-keys/%s", PathSecrets, keystore.SecretName)
			importValidatorContainer := corev1.Container{
				Name:  fmt.Sprintf("import-validator-%s", keystore.SecretName),
				Image: img,
				Args: []string{
					"accounts",
					"import",
					PrysmAcceptTermsOfUse,
					fmt.Sprintf("--%s", validator.Spec.Network),
					PrysmWalletDir,
					fmt.Sprintf("%s/prysm-wallet", PathBlockchainData),
					PrysmKeysDir,
					fmt.Sprintf("%s/keystore-%d.json", keyDir, i),
					PrysmAccountPasswordFile,
					fmt.Sprintf("%s/password.txt", keyDir),
					PrysmWalletPasswordFile,
					fmt.Sprintf("%s/prysm-wallet/prysm-wallet-password.txt", PathSecrets),
				},
				VolumeMounts: mounts,
			}
			initContainers = append(initContainers, importValidatorContainer)
		}
	}

	if validator.Spec.Client == ethereum2v1alpha1.LighthouseClient {
		// TODO: replace with validator account import init container
		// TODO: follow up with lighthouse PR https://github.com/sigp/lighthouse/issues/2224
		validatorsPath := fmt.Sprintf("%s/validators", PathBlockchainData)
		linkValidatorDefinitions := corev1.Container{
			Name:    "link-validator-definitions",
			Image:   "busybox",
			Command: []string{"/bin/sh", "-c"},
			Args: []string{
				fmt.Sprintf("mkdir -p %s && cp %s/validator_definitions.yml %s/validator_definitions.yml && cp -R %s/validator-keys/. %s/validator-keys", validatorsPath, PathConfig, validatorsPath, PathSecrets, PathBlockchainData),
			},
			VolumeMounts: mounts,
		}
		initContainers = append(initContainers, linkValidatorDefinitions)
	}

	sts.Spec = appsv1.StatefulSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: validator.GetLabels(),
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validator.GetLabels(),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:         "validator",
						Image:        img,
						Command:      command,
						Args:         args,
						VolumeMounts: mounts,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(validator.Spec.Resources.CPU),
								corev1.ResourceMemory: resource.MustParse(validator.Spec.Resources.Memory),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(validator.Spec.Resources.CPULimit),
								corev1.ResourceMemory: resource.MustParse(validator.Spec.Resources.MemoryLimit),
							},
						},
					},
				},
				InitContainers: initContainers,
				Volumes:        r.createValidatorVolumes(validator),
			},
		},
	}
}

// reconcileValidatorStatefulset reconciles node statefulset
func (r *ValidatorReconciler) reconcileValidatorStatefulset(validator *ethereum2v1alpha1.Validator) error {
	sts := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      validator.Name,
			Namespace: validator.Namespace,
		},
	}

	client, err := NewValidatorClient(validator.Spec.Client)
	if err != nil {
		return err
	}
	img := client.Image()
	command := client.Command()
	args := client.Args(validator)

	_, err = ctrl.CreateOrUpdate(context.Background(), r.Client, &sts, func() error {
		if err := ctrl.SetControllerReference(validator, &sts, r.Scheme); err != nil {
			return err
		}

		r.specValidatorStatefulset(validator, &sts, img, command, args)

		return nil
	})

	return err
}

// SetupWithManager adds reconciler to the manager
func (r *ValidatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ethereum2v1alpha1.Validator{}).
		Complete(r)
}
