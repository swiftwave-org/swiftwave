package main

import (
	"os"
	"strings"
)

const DataplaneAPIBaseAddress = "http://localhost:3334"

func installHAProxy(swiftwaveAddress string, dnsServer string, username string, password string) error {
	commands := []string{
		"apt remove haproxy -y || true",
		"apt install --no-install-recommends software-properties-common",
		"add-apt-repository ppa:vbernat/haproxy-3.0 -y",
		"apt install haproxy=3.0.* -y",
		"wget -O /tmp/dataplaneapi.deb https://github.com/haproxytech/dataplaneapi/releases/download/v3.0.4/dataplaneapi_3.0.4_linux_" + GetCPUArchitecture() + ".deb",
		"dpkg -i /tmp/dataplaneapi.deb && apt-get install -f -y",
		"rm /tmp/dataplaneapi.deb || true",
		"systemctl daemon-reload",
	}
	for _, command := range commands {
		err := RunCommandWithoutBuffer(command)
		if err != nil {
			return err
		}
	}

	// Create required directories
	os.MkdirAll("/etc/haproxy/maps", 0777)
	os.MkdirAll("/etc/haproxy/ssl", 0777)
	os.MkdirAll("/etc/haproxy/spoe", 0777)

	// Write default.pem TODO: try to generate and add
	err := os.WriteFile("/etc/haproxy/ssl/default.pem", []byte(default_pem_cert), 0777)
	if err != nil {
		return err
	}

	// Write error pages
	err = os.WriteFile("/etc/haproxy/errors/502.http", []byte(http_502), 0777)
	if err != nil {
		return err
	}
	err = os.WriteFile("/etc/haproxy/errors/503.http", []byte(http_503), 0777)
	if err != nil {
		return err
	}

	// Write default haproxy config
	err = os.WriteFile("/etc/haproxy/haproxy.cfg", []byte(haproxyConfig), 0777)
	if err != nil {
		return err
	}

	// Write default dataplaneapi config
	dataplaneapiConfigContent := dataplaneapiConfig
	dataplaneapiConfigContent = strings.ReplaceAll(dataplaneapiConfigContent, "{{ .userID }}", username)
	passwordHash, err := GenerateBasicAuthPassword(password)
	if err != nil {
		return err
	}
	dataplaneapiConfigContent = strings.ReplaceAll(dataplaneapiConfigContent, "{{ .Password }}", passwordHash)
	err = os.WriteFile("/etc/dataplaneapi/dataplaneapi.yml", []byte(dataplaneapiConfigContent), 0777)
	if err != nil {
		return err
	}

	// Write the required variables in /etc/default/haproxy
	envVariables := `
SWIFTWAVE_SERVICE_ENDPOINT="` + swiftwaveAddress + `"
DNS_SERVER="` + dnsServer + `"
`
	err = os.WriteFile("/etc/default/haproxy", []byte(envVariables), 0777)
	if err != nil {
		return err
	}

	// Disable haproxy and data plane api by default
	disableHAProxy()

	return nil
}

func enableHAProxy() {
	_, _, _ = RunCommand("systemctl enable haproxy")
	_, _, _ = RunCommand("systemctl start haproxy")
	_, _, _ = RunCommand("systemctl enable dataplaneapi")
	_, _, _ = RunCommand("systemctl start dataplaneapi")
}

func disableHAProxy() {
	_, _, _ = RunCommand("systemctl stop dataplaneapi")
	_, _, _ = RunCommand("systemctl stop haproxy")
	_, _, _ = RunCommand("systemctl disable dataplaneapi")
	_, _, _ = RunCommand("systemctl disable haproxy")
}

const haproxyConfig = `
global
  master-worker
  maxconn 100000
  chroot /var/lib/haproxy
  user haproxy
  group haproxy
  stats socket /var/run/haproxy.sock user haproxy group haproxy mode 660 level admin expose-fd listeners
  daemon

defaults
  mode http
  option forwardfor
  maxconn 4000
  log global
  option tcp-smart-accept
  timeout http-request 10s
  timeout check 10s
  timeout connect 10s
  timeout client 1m
  timeout queue 1m
  timeout server 1m
  timeout http-keep-alive 10s
  retries 3
  errorfile 502 /etc/haproxy/errors/502.http
  errorfile 503 /etc/haproxy/errors/503.http

resolvers docker
  nameserver ns1 "$DNS_SERVER"
  hold valid    10s
  hold other    30s
  hold refused  30s
  hold nx       30s
  hold timeout  30s
  hold obsolete 30s
  timeout resolve 2s
  timeout retry 2s
  resolve_retries 5
  accepted_payload_size 8192

frontend fe_http
  mode http
  bind :80
  acl lets-encrypt-acl path_beg /.well-known
  use_backend swiftwave_backend if lets-encrypt-acl
  default_backend error_backend

frontend fe_https
  mode http
  bind :443 ssl crt /etc/haproxy/ssl/ alpn h2,http/1.1
  http-request set-header X-Forwarded-Proto https
  acl lets-encrypt-acl path_beg /.well-known
  use_backend swiftwave_backend if lets-encrypt-acl
  default_backend error_backend

frontend fe_swiftwave
  mode http
  bind :3333
  default_backend swiftwave_backend

backend error_backend
  mode http
  http-request deny deny_status 502

backend swiftwave_backend
  option httpchk
  http-check expect status 200
  server swiftwave_service_http "$SWIFTWAVE_SERVICE_ENDPOINT" check

`

var dataplaneapiConfig = `
name: swiftwave_haproxy_dpapi
mode: "single"
dataplaneapi:
  host: "localhost"
  port: 3334
  scheme:
    - http
  user:
  - name: {{ .userID }}
    insecure: false
    password: {{ .Password }}
  resources:
    maps_dir: /etc/haproxy/maps
    ssl_certs_dir: /etc/haproxy/ssl
    spoe_dir: /etc/haproxy/spoe
transaction:
  transaction_dir: "/tmp/haproxy"
  backups_number: 5
  backups_dir: /tmp/backups
  max_open_transactions: 30
haproxy:
  config_file: "/etc/haproxy/haproxy.cfg" # string
  haproxy_bin: "haproxy" # string
  master_runtime: null # string
  fid: null # string
  master_worker_mode: false # bool
  delayed_start_max: 30s # time.Duration
  delayed_start_tick: 500ms # time.Duration
  reload:
    reload_delay: 5
    reload_cmd: "systemctl reload haproxy"
    restart_cmd: "systemctl restart haproxy"
    status_cmd: "systemctl status haproxy"
    service_name: "haproxy.service"
    reload_retention: 1
    reload_strategy: custom
log_targets:
  - log_to: stdout
`

var http_503 = `HTTP/1.1 503 Service Unavailable
Cache-Control: no-cache
Connection: close
Content-Type: text/html

<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="UTF-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<title>503 Service Unavailable</title>
		<style type="text/css">a,hr{color:inherit}progress,sub,sup{vertical-align:baseline}blockquote,body,dd,dl,fieldset,figure,h1,h2,h3,h4,h5,h6,hr,menu,ol,p,pre,ul{margin:0}dialog,fieldset,legend,menu,ol,ul{padding:0}*,::after,::before{box-sizing:border-box;border:0 solid #e5e7eb;--tw-border-spacing-x:0;--tw-border-spacing-y:0;--tw-translate-x:0;--tw-translate-y:0;--tw-rotate:0;--tw-skew-x:0;--tw-skew-y:0;--tw-scale-x:1;--tw-scale-y:1;--tw-pan-x: ;--tw-pan-y: ;--tw-pinch-zoom: ;--tw-scroll-snap-strictness:proximity;--tw-gradient-from-position: ;--tw-gradient-via-position: ;--tw-gradient-to-position: ;--tw-ordinal: ;--tw-slashed-zero: ;--tw-numeric-figure: ;--tw-numeric-spacing: ;--tw-numeric-fraction: ;--tw-ring-inset: ;--tw-ring-offset-width:0px;--tw-ring-offset-color:#fff;--tw-ring-color:rgb(59 130 246 / 0.5);--tw-ring-offset-shadow:0 0 #0000;--tw-ring-shadow:0 0 #0000;--tw-shadow:0 0 #0000;--tw-shadow-colored:0 0 #0000;--tw-blur: ;--tw-brightness: ;--tw-contrast: ;--tw-grayscale: ;--tw-hue-rotate: ;--tw-invert: ;--tw-saturate: ;--tw-sepia: ;--tw-drop-shadow: ;--tw-backdrop-blur: ;--tw-backdrop-brightness: ;--tw-backdrop-contrast: ;--tw-backdrop-grayscale: ;--tw-backdrop-hue-rotate: ;--tw-backdrop-invert: ;--tw-backdrop-opacity: ;--tw-backdrop-saturate: ;--tw-backdrop-sepia: ;--tw-contain-size: ;--tw-contain-layout: ;--tw-contain-paint: ;--tw-contain-style: }::after,::before{--tw-content:''}:host,html{line-height:1.5;-webkit-text-size-adjust:100%;-moz-tab-size:4;tab-size:4;font-family:ui-sans-serif,system-ui,sans-serif,"Apple Color Emoji","Segoe UI Emoji","Segoe UI Symbol","Noto Color Emoji";font-feature-settings:normal;font-variation-settings:normal;-webkit-tap-highlight-color:transparent}body{line-height:inherit}hr{height:0;border-top-width:1px}abbr:where([title]){text-decoration:underline dotted}h1,h2,h3,h4,h5,h6{font-size:inherit;font-weight:inherit}a{text-decoration:inherit}b,strong{font-weight:bolder}code,kbd,pre,samp{font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono","Courier New",monospace;font-feature-settings:normal;font-variation-settings:normal;font-size:1em}small{font-size:80%}sub,sup{font-size:75%;line-height:0;position:relative}sub{bottom:-.25em}sup{top:-.5em}table{text-indent:0;border-color:inherit;border-collapse:collapse}button,input,optgroup,select,textarea{font-family:inherit;font-feature-settings:inherit;font-variation-settings:inherit;font-size:100%;font-weight:inherit;line-height:inherit;letter-spacing:inherit;color:inherit;margin:0;padding:0}button,select{text-transform:none}button,input:where([type=button]),input:where([type=reset]),input:where([type=submit]){-webkit-appearance:button;background-color:transparent;background-image:none}:-moz-focusring{outline:auto}:-moz-ui-invalid{box-shadow:none}::-webkit-inner-spin-button,::-webkit-outer-spin-button{height:auto}[type=search]{-webkit-appearance:textfield;outline-offset:-2px}::-webkit-search-decoration{-webkit-appearance:none}::-webkit-file-upload-button{-webkit-appearance:button;font:inherit}summary{display:list-item}menu,ol,ul{list-style:none}textarea{resize:vertical}input::placeholder,textarea::placeholder{opacity:1;color:#9ca3af}[role=button],button{cursor:pointer}:disabled{cursor:default}audio,canvas,embed,iframe,img,object,svg,video{display:block;vertical-align:middle}img,video{max-width:100%;height:auto}[hidden]{display:none}::backdrop{--tw-border-spacing-x:0;--tw-border-spacing-y:0;--tw-translate-x:0;--tw-translate-y:0;--tw-rotate:0;--tw-skew-x:0;--tw-skew-y:0;--tw-scale-x:1;--tw-scale-y:1;--tw-pan-x: ;--tw-pan-y: ;--tw-pinch-zoom: ;--tw-scroll-snap-strictness:proximity;--tw-gradient-from-position: ;--tw-gradient-via-position: ;--tw-gradient-to-position: ;--tw-ordinal: ;--tw-slashed-zero: ;--tw-numeric-figure: ;--tw-numeric-spacing: ;--tw-numeric-fraction: ;--tw-ring-inset: ;--tw-ring-offset-width:0px;--tw-ring-offset-color:#fff;--tw-ring-color:rgb(59 130 246 / 0.5);--tw-ring-offset-shadow:0 0 #0000;--tw-ring-shadow:0 0 #0000;--tw-shadow:0 0 #0000;--tw-shadow-colored:0 0 #0000;--tw-blur: ;--tw-brightness: ;--tw-contrast: ;--tw-grayscale: ;--tw-hue-rotate: ;--tw-invert: ;--tw-saturate: ;--tw-sepia: ;--tw-drop-shadow: ;--tw-backdrop-blur: ;--tw-backdrop-brightness: ;--tw-backdrop-contrast: ;--tw-backdrop-grayscale: ;--tw-backdrop-hue-rotate: ;--tw-backdrop-invert: ;--tw-backdrop-opacity: ;--tw-backdrop-saturate: ;--tw-backdrop-sepia: ;--tw-contain-size: ;--tw-contain-layout: ;--tw-contain-paint: ;--tw-contain-style: }.flex{display:flex}.h-screen{height:100vh}.w-screen{width:100vw}.max-w-md{max-width:28rem}.flex-col{flex-direction:column}.items-center{align-items:center}.justify-center{justify-content:center}.space-y-4>:not([hidden])~:not([hidden]){--tw-space-y-reverse:0;margin-top:calc(1rem * calc(1 - var(--tw-space-y-reverse)));margin-bottom:calc(1rem * var(--tw-space-y-reverse))}.bg-gray-100{--tw-bg-opacity:1;background-color:rgb(243 244 246 / var(--tw-bg-opacity))}.px-4{padding-left:1rem;padding-right:1rem}.text-center{text-align:center}.text-4xl{font-size:2.25rem;line-height:2.5rem}.text-lg{font-size:1.125rem;line-height:1.75rem}.font-bold{font-weight:700}.tracking-tight{letter-spacing:-.025em}.text-gray-600{--tw-text-opacity:1;color:rgb(75 85 99 / var(--tw-text-opacity))}.text-gray-900{--tw-text-opacity:1;color:rgb(17 24 39 / var(--tw-text-opacity))}@media (min-width:640px){.sm\:text-5xl{font-size:3rem;line-height:1}}</style>
	</head>
	<body>
		<div class="flex h-screen w-screen flex-col items-center justify-center bg-gray-100 px-4">
			<div class="max-w-md space-y-4 text-center">
				<h1 class="text-4xl font-bold tracking-tight text-gray-900 sm:text-5xl">
					Service Unavailable
				</h1>
				<p class="text-lg text-gray-600">
					We're sorry, but our service is temporarily unavailable. Please try
					again later.
				</p>
			</div>
		</div>
	</body>
</html>`

var http_502 = `HTTP/1.1 502 Bad Gateway
Cache-Control: no-cache
Connection: close
Content-Type: text/html

<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="UTF-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<title>502 Bad Gateway</title>
        <style type="text/css">a,hr{color:inherit}progress,sub,sup{vertical-align:baseline}blockquote,body,dd,dl,fieldset,figure,h1,h2,h3,h4,h5,h6,hr,menu,ol,p,pre,ul{margin:0}dialog,fieldset,legend,menu,ol,ul{padding:0}*,::after,::before{box-sizing:border-box;border:0 solid #e5e7eb;--tw-border-spacing-x:0;--tw-border-spacing-y:0;--tw-translate-x:0;--tw-translate-y:0;--tw-rotate:0;--tw-skew-x:0;--tw-skew-y:0;--tw-scale-x:1;--tw-scale-y:1;--tw-pan-x: ;--tw-pan-y: ;--tw-pinch-zoom: ;--tw-scroll-snap-strictness:proximity;--tw-gradient-from-position: ;--tw-gradient-via-position: ;--tw-gradient-to-position: ;--tw-ordinal: ;--tw-slashed-zero: ;--tw-numeric-figure: ;--tw-numeric-spacing: ;--tw-numeric-fraction: ;--tw-ring-inset: ;--tw-ring-offset-width:0px;--tw-ring-offset-color:#fff;--tw-ring-color:rgb(59 130 246 / 0.5);--tw-ring-offset-shadow:0 0 #0000;--tw-ring-shadow:0 0 #0000;--tw-shadow:0 0 #0000;--tw-shadow-colored:0 0 #0000;--tw-blur: ;--tw-brightness: ;--tw-contrast: ;--tw-grayscale: ;--tw-hue-rotate: ;--tw-invert: ;--tw-saturate: ;--tw-sepia: ;--tw-drop-shadow: ;--tw-backdrop-blur: ;--tw-backdrop-brightness: ;--tw-backdrop-contrast: ;--tw-backdrop-grayscale: ;--tw-backdrop-hue-rotate: ;--tw-backdrop-invert: ;--tw-backdrop-opacity: ;--tw-backdrop-saturate: ;--tw-backdrop-sepia: ;--tw-contain-size: ;--tw-contain-layout: ;--tw-contain-paint: ;--tw-contain-style: }::after,::before{--tw-content:''}:host,html{line-height:1.5;-webkit-text-size-adjust:100%;-moz-tab-size:4;tab-size:4;font-family:ui-sans-serif,system-ui,sans-serif,"Apple Color Emoji","Segoe UI Emoji","Segoe UI Symbol","Noto Color Emoji";font-feature-settings:normal;font-variation-settings:normal;-webkit-tap-highlight-color:transparent}body{line-height:inherit}hr{height:0;border-top-width:1px}abbr:where([title]){text-decoration:underline dotted}h1,h2,h3,h4,h5,h6{font-size:inherit;font-weight:inherit}a{text-decoration:inherit}b,strong{font-weight:bolder}code,kbd,pre,samp{font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono","Courier New",monospace;font-feature-settings:normal;font-variation-settings:normal;font-size:1em}small{font-size:80%}sub,sup{font-size:75%;line-height:0;position:relative}sub{bottom:-.25em}sup{top:-.5em}table{text-indent:0;border-color:inherit;border-collapse:collapse}button,input,optgroup,select,textarea{font-family:inherit;font-feature-settings:inherit;font-variation-settings:inherit;font-size:100%;font-weight:inherit;line-height:inherit;letter-spacing:inherit;color:inherit;margin:0;padding:0}button,select{text-transform:none}button,input:where([type=button]),input:where([type=reset]),input:where([type=submit]){-webkit-appearance:button;background-color:transparent;background-image:none}:-moz-focusring{outline:auto}:-moz-ui-invalid{box-shadow:none}::-webkit-inner-spin-button,::-webkit-outer-spin-button{height:auto}[type=search]{-webkit-appearance:textfield;outline-offset:-2px}::-webkit-search-decoration{-webkit-appearance:none}::-webkit-file-upload-button{-webkit-appearance:button;font:inherit}summary{display:list-item}menu,ol,ul{list-style:none}textarea{resize:vertical}input::placeholder,textarea::placeholder{opacity:1;color:#9ca3af}[role=button],button{cursor:pointer}:disabled{cursor:default}audio,canvas,embed,iframe,img,object,svg,video{display:block;vertical-align:middle}img,video{max-width:100%;height:auto}[hidden]{display:none}::backdrop{--tw-border-spacing-x:0;--tw-border-spacing-y:0;--tw-translate-x:0;--tw-translate-y:0;--tw-rotate:0;--tw-skew-x:0;--tw-skew-y:0;--tw-scale-x:1;--tw-scale-y:1;--tw-pan-x: ;--tw-pan-y: ;--tw-pinch-zoom: ;--tw-scroll-snap-strictness:proximity;--tw-gradient-from-position: ;--tw-gradient-via-position: ;--tw-gradient-to-position: ;--tw-ordinal: ;--tw-slashed-zero: ;--tw-numeric-figure: ;--tw-numeric-spacing: ;--tw-numeric-fraction: ;--tw-ring-inset: ;--tw-ring-offset-width:0px;--tw-ring-offset-color:#fff;--tw-ring-color:rgb(59 130 246 / 0.5);--tw-ring-offset-shadow:0 0 #0000;--tw-ring-shadow:0 0 #0000;--tw-shadow:0 0 #0000;--tw-shadow-colored:0 0 #0000;--tw-blur: ;--tw-brightness: ;--tw-contrast: ;--tw-grayscale: ;--tw-hue-rotate: ;--tw-invert: ;--tw-saturate: ;--tw-sepia: ;--tw-drop-shadow: ;--tw-backdrop-blur: ;--tw-backdrop-brightness: ;--tw-backdrop-contrast: ;--tw-backdrop-grayscale: ;--tw-backdrop-hue-rotate: ;--tw-backdrop-invert: ;--tw-backdrop-opacity: ;--tw-backdrop-saturate: ;--tw-backdrop-sepia: ;--tw-contain-size: ;--tw-contain-layout: ;--tw-contain-paint: ;--tw-contain-style: }.flex{display:flex}.h-screen{height:100vh}.w-full{width:100%}.max-w-md{max-width:28rem}.flex-col{flex-direction:column}.items-center{align-items:center}.justify-center{justify-content:center}.space-y-4>:not([hidden])~:not([hidden]){--tw-space-y-reverse:0;margin-top:calc(1rem * calc(1 - var(--tw-space-y-reverse)));margin-bottom:calc(1rem * var(--tw-space-y-reverse))}.bg-gray-100{--tw-bg-opacity:1;background-color:rgb(243 244 246 / var(--tw-bg-opacity))}.px-4{padding-left:1rem;padding-right:1rem}.text-center{text-align:center}.text-3xl{font-size:1.875rem;line-height:2.25rem}.text-8xl{font-size:6rem;line-height:1}.font-bold{font-weight:700}.font-semibold{font-weight:600}.text-gray-600{--tw-text-opacity:1;color:rgb(75 85 99 / var(--tw-text-opacity))}.text-gray-900{--tw-text-opacity:1;color:rgb(17 24 39 / var(--tw-text-opacity))}</style>
    </head>
	<body>
        <div class="flex h-screen w-full flex-col items-center justify-center bg-gray-100 px-4">
            <div class="max-w-md space-y-4 text-center">
              <h1 class="text-8xl font-bold text-gray-900">502</h1>
              <h2 class="text-3xl font-semibold text-gray-900">Bad Gateway</h2>
              <p class="text-gray-600">
                Looks like you have lost the path to service. Nothing is available here at the moment.
              </p>
            </div>
        </div>
	</body>
</html>`

var default_pem_cert = `
-----BEGIN CERTIFICATE-----
MIICoTCCAYkCFG4t0nUm5HNxekj/pH5/dZ/oUdhLMA0GCSqGSIb3DQEBCwUAMA0x
CzAJBgNVBAYTAlhYMB4XDTIzMTEwODE5MzEwMloXDTI0MTEwNzE5MzEwMlowDTEL
MAkGA1UEBhMCWFgwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC5luRh
Oa+R/yGu3wzB7FEBMB+pb8G91kiyeIKFn1AS+FrU0BmUPB0JfzfKHljGFCKsrYn3
MnijA6eIRzoWM78Rexe/RrnxHtdoLOQH/it4KiQ/mlJ6gsUVzqVnBX+VSYx/4+kE
4lvdS1Z8SCSJq2Mx5KVV+lyNwt9s5Zal8fsH+gOgJyzPKrvABFeBZ2MAvY1JBkgB
udDrzw7gvhMWUDnOsMWJtbR3w2wZ9vSzK+gn8yHDCIRFdnTFssz6byporv3mjCgh
Ln/xmks/WBKvNVGpTnfN0U1URzv4PmALueV+vAiAB+ji9kwQUoqXSAZRC4l6JKjU
CeNxgaIW7LC+8ojzAgMBAAEwDQYJKoZIhvcNAQELBQADggEBADKxBzxlQjDKBpl1
Zv2aykYu8RmoAINY6kpVvQiCbohV7WNdU3NlBxjzUo10o+OrwojyxoqVK2jLbKmo
X0P7o0Eh2IKoIcNjoK1ngBijCajW9rlp0eH932qYIkwABui7SY3bjk9KI6MMso1F
kn7juvaZlbIVXR6b1EwhImOa9aRWKzD9b6G03YfCDGbraMNyPAj/g5cpyOGAtlts
pJ7sbRBvhtOQJ68jBOLMz8rBp8H7jAL+LaibSsqQC2o+XtWnS1WgYvryh0g2IDYw
eCXs0vR86rQF4YdRMmsr2t3+acF9/TJM9PtHhTiz+ioElQiPvmU+c3CRn6EqbKmg
/18KFTo=
-----END CERTIFICATE-----
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC5luRhOa+R/yGu
3wzB7FEBMB+pb8G91kiyeIKFn1AS+FrU0BmUPB0JfzfKHljGFCKsrYn3MnijA6eI
RzoWM78Rexe/RrnxHtdoLOQH/it4KiQ/mlJ6gsUVzqVnBX+VSYx/4+kE4lvdS1Z8
SCSJq2Mx5KVV+lyNwt9s5Zal8fsH+gOgJyzPKrvABFeBZ2MAvY1JBkgBudDrzw7g
vhMWUDnOsMWJtbR3w2wZ9vSzK+gn8yHDCIRFdnTFssz6byporv3mjCghLn/xmks/
WBKvNVGpTnfN0U1URzv4PmALueV+vAiAB+ji9kwQUoqXSAZRC4l6JKjUCeNxgaIW
7LC+8ojzAgMBAAECggEAAYERaR+nESIGELKab13uICTW34WE2/gxzvnUEPzZZiCg
19W3R1t0tCArAGveNf4cHoB/kZh/9Xjw8X/7nrTHqSlxZ/73ldaKp2ZUaM4suwTD
Fh2MP+rxvvtVnYVOuJPdMXzUOtJngoklnSHr4zkXjOQjWj7fuNRqAX9w3k0c8ZPC
bjPnNK36qICAyD0k1z23ppxSy3tq1+opnnf+Yl9WGRsAxxkjcXRoTgCVZRTTGC+F
sK0zkyJwgpm4ayV3ckM+TF9csxO876k7uPf4WFzVm5oEDGvx75h5w8Od35NUgkRA
0evz1YL/Q4N8lASXOTuam3Fq1gQC04zwpRdGeHjadQKBgQC70fQ6hx1SHMvahag4
h0qcQzpyBBfEismfi7QMKz04TyCb7M+s8Bd6gkho8ON9iww5uuzLPsoieRDgtikV
M213uZeIecoiNpxILo0ulRq5RqsxvqOmL2T0JM6sYSUCm3DRwr0MhbQNFcSFVWhl
GDYk/hfQvr8RF8LewjtIX6b3TQKBgQD89aObpljBrzEfNNBlz4fg4S4KS+3hUx80
EYK/fguhS5+kOO7D2a0vKn3q8krm8//ynogKzAcuX8y6SaFzFE+h1RThT03eTRq+
nE6OJzDcIZgdw/74vGwG96Ptkb42s7HUQWDTyWEJ9UsQgDSoxwOTsZ2y2BO2sEn0
ZTLIpQXhPwKBgQCp9E9i0rbGgcY5Y+6X0FzET9VILMnxEIFn/Lucs1e/Z2KjlcNK
wysLsW6ify/rf3I9nxb8x0GTtid+n3dHdvTcfLVRSpuNIAuFCZK5jzTSaM8qwU5G
Z+abQd8+ft1FobCSLvxwo2AM4yCkYmeH60O7b63PN3uflPfCKNIKKHvmlQKBgBz6
tyeZwwlNXL9KeaVwRQzKP1AGqtXpg+WfK+9sLUDpPPy/WPsu8Nw6bfqAj3wt7+CH
sOYrwZbaesXMsaZRaV4M3zuArlcNVkcH+Sfn7X0KjDa8wXUVgPq7XBhXXgc+Rt0e
ME2TAH73jwXw6hd71TkSXBKlFn0TbSWGgm7iGO5ZAoGAc8MWMtjBEQYFCF6LpYbs
Gjsny1hfqvX9KOkjGxrvlmd0D6yXRwL7ESBtE/ugVpP14mrVFX1653FlpDAGBNmU
TPYZiWnJ2nnuKJDDR4sY+Tx9KKkGvVwUwwaaBKL0VjMk7ORABiLGwL+byVriKLnN
5wA0oRrlSszpHHDy0fhdQMI=
-----END PRIVATE KEY-----
`
