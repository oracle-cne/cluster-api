/*
Copyright 2024 The Kubernetes Authors.

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

package fcos1_5

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"text/template"

	"github.com/coreos/butane/config"
	butanecommon "github.com/coreos/butane/config/common"
	"github.com/coreos/ignition/v2/config/util"
	"github.com/coreos/ignition/v2/config/v3_4"
	"github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/coreos/vcontext/report"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/internal/cloudinit"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/internal/ignition/common"
)

const (
	butaneTemplate = `---
variant: fcos
version: 1.5.0
{{- if .Users }}
passwd:
  users:
    {{- range .Users }}
    - name: {{ .Name }}
      {{- with .Gecos }}
      gecos: {{ . }}
      {{- end }}
      {{- if .Groups }}
      groups:
        {{- range Split .Groups ", " }}
        - {{ . }}
        {{- end }}
      {{- end }}
      {{- with .HomeDir }}
      home_dir: {{ . }}
      {{- end }}
      {{- with .Shell }}
      shell: {{ . }}
      {{- end }}
      {{- with .Passwd }}
      password_hash: {{ . }}
      {{- end }}
      {{- with .PrimaryGroup }}
      primary_group: {{ . }}
      {{- end }}
      {{- if .SSHAuthorizedKeys }}
      ssh_authorized_keys:
        {{- range .SSHAuthorizedKeys }}
        - {{ . }}
        {{- end }}
      {{- end }}
    {{- end }}
{{- end }}
systemd:
  units:
    - name: kubeadm.service
      enabled: true
      contents: |
        [Unit]
        Description=kubeadm
        # Run only once. After successful run, this file is moved to /tmp/.
        ConditionPathExists=/etc/kubeadm.yml
        After=network.target
        [Service]
        # To not restart the unit when it exits, as it is expected.
        Type=oneshot
        ExecStart=/etc/kubeadm.sh
        [Install]
        WantedBy=multi-user.target
    {{- if .NTP }}{{ if .NTP.Enabled }}
    - name: ntpd.service
      enabled: true
    {{- end }}{{- end }}
    {{- range .Mounts }}
    {{- $label := index . 0 }}
    {{- $mountpoint := index . 1 }}
    {{- $disk := index $.FilesystemDevicesByLabel $label }}
    {{- $mountOptions := slice . 2 }}
    - name: {{ $mountpoint | MountpointName }}.mount
      enabled: true
      contents: |
        [Unit]
        Description = Mount {{ $label }}

        [Mount]
        What={{ $disk }}
        Where={{ $mountpoint }}
        Options={{ Join $mountOptions "," }}

        [Install]
        WantedBy=multi-user.target
    {{- end }}
storage:
  {{- if .DiskSetup }}{{- if .DiskSetup.Partitions }}
  disks:
    {{- range .DiskSetup.Partitions }}
    - device: {{ .Device }}
      {{- with .Overwrite }}
      wipe_table: {{ . }}
      {{- end }}
      {{- if .Layout }}
      partitions:
      - {}
      {{- end }}
    {{- end }}
  {{- end }}{{- end }}
  {{- if .DiskSetup }}{{- if .DiskSetup.Filesystems }}
  filesystems:
    {{- range .DiskSetup.Filesystems }}
    - device: {{ .Device }}
      format: {{ .Filesystem }}
      wipe_filesystem: {{ .Overwrite }}
      label: {{ .Label }}
      {{- if .ExtraOpts }}
      options:
        {{- range .ExtraOpts }}
        - {{ . }}
        {{- end }}
      {{- end }}
    {{- end }}
  {{- end }}{{- end }}
  files:
    {{- range .Users }}
    {{- if .Sudo }}
    - path: /etc/sudoers.d/{{ .Name }}
      mode: 0600
      contents:
        inline: |
          {{ .Name }} {{ .Sudo }}
    {{- end }}
    {{- end }}
    {{- with .UsersWithPasswordAuth }}
    - path: /etc/ssh/sshd_config
      mode: 0600
      contents:
        inline: |
          # Use most defaults for sshd configuration.
          Subsystem sftp internal-sftp
          ClientAliveInterval 180
          UseDNS no
          UsePAM yes
          PrintLastLog no # handled by PAM
          PrintMotd no # handled by PAM

          Match User {{ . }}
            PasswordAuthentication yes
    {{- end }}
    {{- range .WriteFiles }}
    - path: {{ .Path }}
      {{- $owner := ParseOwner .Owner }}
      {{ if $owner.User -}}
      user:
        name: {{ $owner.User }}
      {{- end }}
      {{ if $owner.Group -}}
      group:
        name: {{ $owner.Group }}
      {{- end }}
      # Owner
      {{ if ne .Permissions "" -}}
      mode: {{ .Permissions }}
      {{ end -}}
      contents:
        {{ if eq .Encoding "base64" -}}
        inline: !!binary |
        {{- else -}}
        inline: |
        {{- end }}
          {{ .Content | Indent 10 }}
    {{- end }}
    - path: /etc/kubeadm.sh
      mode: 0700
      contents:
        inline: |
          #!/bin/bash
          set -e
          {{ range .PreKubeadmCommands }}
          {{ . | Indent 10 }}
          {{- end }}

          {{ .KubeadmCommand }}
          mkdir -p /run/cluster-api && echo success > /run/cluster-api/bootstrap-success.complete
          mv /etc/kubeadm.yml /tmp/
          {{range .PostKubeadmCommands }}
          {{ . | Indent 10 }}
          {{- end }}
    - path: /etc/kubeadm.yml
      mode: 0600
      contents:
        inline: |
          ---
          {{ .KubeadmConfig | Indent 10 }}
    {{- if .NTP }}{{- if and .NTP.Enabled .NTP.Servers }}
    - path: /etc/ntp.conf
      mode: 0644
      contents:
        inline: |
          # Common pool
          {{- range  .NTP.Servers }}
          server {{ . }}
          {{- end }}

          # Warning: Using default NTP settings will leave your NTP
          # server accessible to all hosts on the Internet.

          # If you want to deny all machines (including your own)
          # from accessing the NTP server, uncomment:
          #restrict default ignore

          # Default configuration:
          # - Allow only time queries, at a limited rate, sending KoD when in excess.
          # - Allow all local queries (IPv4, IPv6)
          restrict default nomodify nopeer noquery notrap limited kod
          restrict 127.0.0.1
          restrict [::1]
    {{- end }}{{- end }}
`
    additionalConfigTemplate = `---
variant: fcos
version: 1.5.0
{{ .ContainerLinuxConfig.AdditionalConfig }}
`
)

func butaneToIgnition(input []byte) (types.Config, report.Report, error) {
	// Convert the butane bytes into ignition bytes, and then parse those
	// into the correction ignition Config struct.  This is not especially
	// efficient.  However, this is not a hot path, and it avoids duplicating
	// significant portions of the Butane code.
	ignBytes, report, err := config.TranslateBytes(input, butanecommon.TranslateBytesOptions{
		Raw: true,
	})
	if err != nil {
		return types.Config{}, report, err
	}
	ignConfig, report, err := v3_4.ParseCompatibleVersion(ignBytes)
	if err != nil {
		return types.Config{}, report, err
	}
	return ignConfig, report, nil
}

func Render(input *cloudinit.BaseUserData, clc *bootstrapv1.ContainerLinuxConfig, kubeadmConfig string) ([]byte, string, error) {
	if input == nil {
		return nil, "", errors.New("empty base user data")
	}

	t := template.Must(template.New("template").Funcs(common.DefaultTemplateFunctions()).Parse(butaneTemplate))

	usersWithPasswordAuth := []string{}
	for _, user := range input.Users {
		if user.LockPassword != nil && !*user.LockPassword {
			usersWithPasswordAuth = append(usersWithPasswordAuth, user.Name)
		}
	}

	filesystemDevicesByLabel := map[string]string{}
	if input.DiskSetup != nil {
		for _, filesystem := range input.DiskSetup.Filesystems {
			filesystemDevicesByLabel[filesystem.Label] = filesystem.Device
		}
	}

	// Make sure that clc is not nil to simplify the template
	if clc == nil {
		clc = &bootstrapv1.ContainerLinuxConfig{
		}
	}

	data := common.Data{
		BaseUserData: input,
		KubeadmConfig: kubeadmConfig,
		UsersWithPasswordAuth: strings.Join(usersWithPasswordAuth, ","),
		FilesystemDevicesByLabel: filesystemDevicesByLabel,
		ContainerLinuxConfig: clc,
	}

	var out bytes.Buffer
	if err := t.Execute(&out, data); err != nil {
		return nil, "", err
	}

	ignConfig, reports, err := butaneToIgnition(out.Bytes())
	if err != nil {
		return nil, reports.String(), err
	}
	if clc.AdditionalConfig != "" {
		// Tolerate ignition as well as butane.  If the input is ignition, just
		// return the parsed values.  If not, treat the input as thought it were
		// supposed to be butane.
		//
		// Ignition errors are only respected if it is valid enough to have a
		// discernable version.  If not, drop into the butane handler.
		var addlIgnConfig types.Config
		var addlReport report.Report
		ver, _, _ := util.GetConfigVersion([]byte(clc.AdditionalConfig))
		if ver.String() != "0.0.0" {
			addlIgnConfig, addlReport, err = v3_4.ParseCompatibleVersion([]byte(clc.AdditionalConfig))
			if err != nil {
				return nil, addlReport.String(), err
			}
		} else {
			t := template.Must(template.New("template").Funcs(common.DefaultTemplateFunctions()).Parse(additionalConfigTemplate))
			var out bytes.Buffer
			err = t.Execute(&out, data)
			if err != nil {
				return nil, "", err
			}

			addlIgnConfig, addlReport, err = butaneToIgnition(out.Bytes())
			if err != nil {
				return nil, addlReport.String(), err
			}
		}

		merged := v3_4.Merge(ignConfig, addlIgnConfig)
		ignConfig = merged
		reports.Merge(addlReport)
	}

	ignBytes, err := json.Marshal(&ignConfig)
	if err != nil {
		return nil, reports.String(), err
	}
	return ignBytes, reports.String(), err
}

