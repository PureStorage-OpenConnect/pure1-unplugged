{{/*
Get the FQDN from the global.publicAddress value. If its an IP address set ""
*/}}
{{- define "publicAddress.domainName" -}}
{{- if not (regexMatch "^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$" .Values.global.publicAddress) -}}
{{- .Values.global.publicAddress -}}
{{- end -}}
{{- end -}}
