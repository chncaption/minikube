/*
Copyright 2019 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cruntime

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"path"
	"strings"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

const crioConfigTemplate = `# The CRI-O configuration file specifies all of the available configuration
# options and command-line flags for the crio(8) OCI Kubernetes Container Runtime
# daemon, but in a TOML format that can be more easily modified and versioned.
#
# Please refer to crio.conf(5) for details of all configuration options.

# CRI-O supports partial configuration reload during runtime, which can be
# done by sending SIGHUP to the running process. Currently supported options
# are explicitly mentioned with: 'This option supports live configuration
# reload'.

# CRI-O reads its storage defaults from the containers-storage.conf(5) file
# located at /etc/containers/storage.conf. Modify this storage configuration if
# you want to change the system's defaults. If you want to modify storage just
# for CRI-O, you can change the storage configuration options here.
[crio]

# Path to the "root directory". CRI-O stores all of its data, including
# containers images, in this directory.
root = "/var/lib/containers/storage"

# Path to the "run directory". CRI-O stores all of its state in this directory.
runroot = "/var/run/containers/storage"

# Storage driver used to manage the storage of images and containers. Please
# refer to containers-storage.conf(5) to see all available storage drivers.
storage_driver = "overlay"

# List to pass options to the storage driver. Please refer to
# containers-storage.conf(5) to see all available storage options.
#storage_option = [
#]

# If set to false, in-memory locking will be used instead of file-based locking.
# **Deprecated** this option will be removed in the future.
file_locking = false

# Path to the lock file.
# **Deprecated** this option will be removed in the future.
file_locking_path = "/run/crio.lock"


# The crio.api table contains settings for the kubelet/gRPC interface.
[crio.api]

# Path to AF_LOCAL socket on which CRI-O will listen.
listen = "/var/run/crio/crio.sock"

# IP address on which the stream server will listen.
stream_address = "127.0.0.1"

# The port on which the stream server will listen.
stream_port = "0"

# Enable encrypted TLS transport of the stream server.
stream_enable_tls = false

# Path to the x509 certificate file used to serve the encrypted stream. This
# file can change, and CRI-O will automatically pick up the changes within 5
# minutes.
stream_tls_cert = ""

# Path to the key file used to serve the encrypted stream. This file can
# change, and CRI-O will automatically pick up the changes within 5 minutes.
stream_tls_key = ""

# Path to the x509 CA(s) file used to verify and authenticate client
# communication with the encrypted stream. This file can change, and CRI-O will
# automatically pick up the changes within 5 minutes.
stream_tls_ca = ""

# Maximum grpc send message size in bytes. If not set or <=0, then CRI-O will default to 16 * 1024 * 1024.
grpc_max_send_msg_size = 16777216

# Maximum grpc receive message size. If not set or <= 0, then CRI-O will default to 16 * 1024 * 1024.
grpc_max_recv_msg_size = 16777216

# The crio.runtime table contains settings pertaining to the OCI runtime used
# and options for how to set up and manage the OCI runtime.
[crio.runtime]

# A list of ulimits to be set in containers by default, specified as
# "<ulimit name>=<soft limit>:<hard limit>", for example:
# "nofile=1024:2048"
# If nothing is set here, settings will be inherited from the CRI-O daemon
#default_ulimits = [
#]

# default_runtime is the _name_ of the OCI runtime to be used as the default.
# The name is matched against the runtimes map below.
default_runtime = "runc"

# If true, the runtime will not use pivot_root, but instead use MS_MOVE.
no_pivot = true

# Path to the conmon binary, used for monitoring the OCI runtime.
conmon = "/usr/libexec/crio/conmon"

# Cgroup setting for conmon
conmon_cgroup = "pod"

# Environment variable list for the conmon process, used for passing necessary
# environment variables to conmon or the runtime.
conmon_env = [
	"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
]

# If true, SELinux will be used for pod separation on the host.
selinux = false

# Path to the seccomp.json profile which is used as the default seccomp profile
# for the runtime. If not specified, then the internal default seccomp profile
# will be used.
seccomp_profile = ""

# Used to change the name of the default AppArmor profile of CRI-O. The default
# profile name is "crio-default-" followed by the version string of CRI-O.
apparmor_profile = "crio-default"

# Cgroup management implementation used for the runtime.
cgroup_manager = "cgroupfs"

# List of default capabilities for containers. If it is empty or commented out,
# only the capabilities defined in the containers json file by the user/kube
# will be added.
default_capabilities = [
	"CHOWN",
	"DAC_OVERRIDE",
	"FSETID",
	"FOWNER",
	"NET_RAW",
	"SETGID",
	"SETUID",
	"SETPCAP",
	"NET_BIND_SERVICE",
	"SYS_CHROOT",
	"KILL",
]

# List of default sysctls. If it is empty or commented out, only the sysctls
# defined in the container json file by the user/kube will be added.
default_sysctls = [
]

# List of additional devices. specified as
# "<device-on-host>:<device-on-container>:<permissions>", for example: "--device=/dev/sdc:/dev/xvdc:rwm".
#If it is empty or commented out, only the devices
# defined in the container json file by the user/kube will be added.
additional_devices = [
]

# Path to OCI hooks directories for automatically executed hooks.
hooks_dir = [
]

# List of default mounts for each container. **Deprecated:** this option will
# be removed in future versions in favor of default_mounts_file.
default_mounts = [
]

# Path to the file specifying the defaults mounts for each container. The
# format of the config is /SRC:/DST, one mount per line. Notice that CRI-O reads
# its default mounts from the following two files:
#
#   1) /etc/containers/mounts.conf (i.e., default_mounts_file): This is the
#      override file, where users can either add in their own default mounts, or
#      override the default mounts shipped with the package.
#
#   2) /usr/share/containers/mounts.conf: This is the default file read for
#      mounts. If you want CRI-O to read from a different, specific mounts file,
#      you can change the default_mounts_file. Note, if this is done, CRI-O will
#      only add mounts it finds in this file.
#
#default_mounts_file = ""

# Maximum number of processes allowed in a container.
pids_limit = 1024

# Maximum sized allowed for the container log file. Negative numbers indicate
# that no size limit is imposed. If it is positive, it must be >= 8192 to
# match/exceed conmon's read buffer. The file is truncated and re-opened so the
# limit is never exceeded.
log_size_max = -1

# Whether container output should be logged to journald in addition to the kuberentes log file
log_to_journald = false

# Path to directory in which container exit files are written to by conmon.
container_exits_dir = "/var/run/crio/exits"

# Path to directory for container attach sockets.
container_attach_socket_dir = "/var/run/crio"

# If set to true, all containers will run in read-only mode.
read_only = false

# Changes the verbosity of the logs based on the level it is set to. Options
# are fatal, panic, error, warn, info, and debug. This option supports live
# configuration reload.
log_level = "error"

# The default log directory where all logs will go unless directly specified by the kubelet
log_dir = "/var/log/crio/pods"

# The UID mappings for the user namespace of each container. A range is
# specified in the form containerUID:HostUID:Size. Multiple ranges must be
# separated by comma.
uid_mappings = ""

# The GID mappings for the user namespace of each container. A range is
# specified in the form containerGID:HostGID:Size. Multiple ranges must be
# separated by comma.
gid_mappings = ""

# The minimal amount of time in seconds to wait before issuing a timeout
# regarding the proper termination of the container.
ctr_stop_timeout = 0

# ManageNetworkNSLifecycle determines whether we pin and remove network namespace
# and manage its lifecycle.
manage_network_ns_lifecycle = false

# The "crio.runtime.runtimes" table defines a list of OCI compatible runtimes.
# The runtime to use is picked based on the runtime_handler provided by the CRI.
# If no runtime_handler is provided, the runtime will be picked based on the level
# of trust of the workload.

[crio.runtime.runtimes.runc]
runtime_path = "/usr/bin/runc"
runtime_type = "oci"
runtime_root = "/run/runc"


# The crio.image table contains settings pertaining to the management of OCI images.
#
# CRI-O reads its configured registries defaults from the system wide
# containers-registries.conf(5) located in /etc/containers/registries.conf. If
# you want to modify just CRI-O, you can change the registries configuration in
# this file. Otherwise, leave insecure_registries and registries commented out to
# use the system's defaults from /etc/containers/registries.conf.
[crio.image]

# Default transport for pulling images from a remote container storage.
default_transport = "docker://"

# The path to a file containing credentials necessary for pulling images from
# secure registries. The file is similar to that of /var/lib/kubelet/config.json
global_auth_file = ""

# The image used to instantiate infra containers.
# This option supports live configuration reload.
pause_image = "{{ .PodInfraContainerImage }}"

# The path to a file containing credentials specific for pulling the pause_image from
# above. The file is similar to that of /var/lib/kubelet/config.json
# This option supports live configuration reload.
pause_image_auth_file = ""

# The command to run to have a container stay in the paused state.
# This option supports live configuration reload.
pause_command = "/pause"

# Path to the file which decides what sort of policy we use when deciding
# whether or not to trust an image that we've pulled. It is not recommended that
# this option be used, as the default behavior of using the system-wide default
# policy (i.e., /etc/containers/policy.json) is most often preferred. Please
# refer to containers-policy.json(5) for more details.
signature_policy = ""

# Controls how image volumes are handled. The valid values are mkdir, bind and
# ignore; the latter will ignore volumes entirely.
image_volumes = "mkdir"

# List of registries to be used when pulling an unqualified image (e.g.,
# "alpine:latest"). By default, registries is set to "docker.io" for
# compatibility reasons. Depending on your workload and usecase you may add more
# registries (e.g., "quay.io", "registry.fedoraproject.org",
# "registry.opensuse.org", etc.).
registries = [
	"docker.io"
]


# The crio.network table containers settings pertaining to the management of
# CNI plugins.
[crio.network]

# Path to the directory where CNI configuration files are located.
network_dir = "/etc/cni/net.d/"

# Paths to directories where CNI plugin binaries are located.
plugin_dirs = [
	"/opt/cni/bin/",
]
`

// listCRIContainers returns a list of containers using crictl
func listCRIContainers(cr CommandRunner, filter string) ([]string, error) {
	var content string
	var err error
	state := "Running"
	if filter != "" {
		content, err = cr.CombinedOutput(fmt.Sprintf(`sudo crictl ps -a --name=%s --state=%s --quiet`, filter, state))
	} else {
		content, err = cr.CombinedOutput(fmt.Sprintf(`sudo crictl ps -a --state=%s --quiet`, state))
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, line := range strings.Split(content, "\n") {
		if line != "" {
			ids = append(ids, line)
		}
	}
	return ids, nil
}

// criCRIContainers kills a list of containers using crictl
func killCRIContainers(cr CommandRunner, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	glog.Infof("Killing containers: %s", ids)
	return cr.Run(fmt.Sprintf("sudo crictl rm %s", strings.Join(ids, " ")))
}

// stopCRIContainers stops containers using crictl
func stopCRIContainers(cr CommandRunner, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	glog.Infof("Stopping containers: %s", ids)
	return cr.Run(fmt.Sprintf("sudo crictl stop %s", strings.Join(ids, " ")))
}

// populateCRIConfig sets up /etc/crictl.yaml
func populateCRIConfig(cr CommandRunner, socket string) error {
	cPath := "/etc/crictl.yaml"
	tmpl := `runtime-endpoint: unix://{{.Socket}}
image-endpoint: unix://{{.Socket}}
`
	t, err := template.New("crictl").Parse(tmpl)
	if err != nil {
		return err
	}
	opts := struct{ Socket string }{Socket: socket}
	var b bytes.Buffer
	if err := t.Execute(&b, opts); err != nil {
		return err
	}
	return cr.Run(fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s", path.Dir(cPath), b.String(), cPath))
}

// generateCRIOConfig sets up /etc/crio/crio.conf
func generateCRIOConfig(cr CommandRunner, k8s config.KubernetesConfig) error {
	cPath := constants.CRIOConfFile
	t, err := template.New("crio.conf").Parse(crioConfigTemplate)
	if err != nil {
		return err
	}
	podInfraContainerImage, _ := images.CachedImages(k8s.ImageRepository, k8s.KubernetesVersion)
	opts := struct{ PodInfraContainerImage string }{PodInfraContainerImage: podInfraContainerImage}
	var b bytes.Buffer
	if err := t.Execute(&b, opts); err != nil {
		return err
	}
	return cr.Run(fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | base64 -d | sudo tee %s", path.Dir(cPath), base64.StdEncoding.EncodeToString(b.Bytes()), cPath))
}

// criContainerLogCmd returns the command to retrieve the log for a container based on ID
func criContainerLogCmd(id string, len int, follow bool) string {
	var cmd strings.Builder
	cmd.WriteString("sudo crictl logs ")
	if len > 0 {
		cmd.WriteString(fmt.Sprintf("--tail %d ", len))
	}
	if follow {
		cmd.WriteString("--follow ")
	}

	cmd.WriteString(id)
	return cmd.String()
}
