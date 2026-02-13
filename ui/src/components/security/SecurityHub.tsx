import { useEffect, useState } from "react";
import { useMTLSStore } from "@/stores/mtlsStore";
import { useSecurityStore } from "@/stores/securityStore";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Loader2,
  Shield,
  ShieldCheck,
  ShieldAlert,
  AlertTriangle,
  Info,
  CheckCircle,
  XCircle,
  Plug,
  Lock,
  FileKey,
} from "lucide-react";
import { CAManager } from "./CAManager";
import { ClientCertList } from "./ClientCertList";
import { CertImportGuide } from "./CertImportGuide";
import type { SecureHeadersConfig } from "@/types";

export function SecurityHub() {
  const {
    config,
    loading,
    error,
    pluginStatus,
    middlewareConfig,
    fetchConfig,
    enableMTLS,
    disableMTLS,
    checkPlugin,
    fetchMiddlewareConfig,
    updateMiddlewareConfig,
    clearError,
  } = useMTLSStore();

  const {
    config: securityConfig,
    error: securityError,
    fetchConfig: fetchSecurityConfig,
    enableTLSHardening,
    disableTLSHardening,
    enableSecureHeaders,
    disableSecureHeaders,
    updateSecureHeadersConfig,
    clearError: clearSecurityError,
  } = useSecurityStore();

  const [showSetupWarning, setShowSetupWarning] = useState(false);
  const [switchLoading, setSwitchLoading] = useState(false);
  const [tlsHardeningLoading, setTLSHardeningLoading] = useState(false);
  const [secureHeadersLoading, setSecureHeadersLoading] = useState(false);
  const [middlewareForm, setMiddlewareForm] = useState({
    rules: "",
    request_headers: "",
    reject_message: "Access denied: Valid client certificate required",
    refresh_interval: 300,
  });
  const [secureHeadersForm, setSecureHeadersForm] = useState<SecureHeadersConfig>({
    x_content_type_options: "nosniff",
    x_frame_options: "SAMEORIGIN",
    x_xss_protection: "1; mode=block",
    hsts: "max-age=31536000; includeSubDomains",
    referrer_policy: "strict-origin-when-cross-origin",
    csp: "",
    permissions_policy: "",
  });
  const [savingMiddleware, setSavingMiddleware] = useState(false);
  const [savingSecureHeaders, setSavingSecureHeaders] = useState(false);

  useEffect(() => {
    fetchConfig();
    checkPlugin();
    fetchMiddlewareConfig();
    fetchSecurityConfig();
  }, [fetchConfig, checkPlugin, fetchMiddlewareConfig, fetchSecurityConfig]);

  useEffect(() => {
    if (middlewareConfig) {
      setMiddlewareForm({
        rules: middlewareConfig.rules || "",
        request_headers: middlewareConfig.request_headers || "",
        reject_message:
          middlewareConfig.reject_message ||
          "Access denied: Valid client certificate required",
        refresh_interval: middlewareConfig.refresh_interval || 300,
      });
    }
  }, [middlewareConfig]);

  useEffect(() => {
    if (securityConfig?.secure_headers) {
      setSecureHeadersForm(securityConfig.secure_headers);
    }
  }, [securityConfig]);

  const handleToggleMTLS = async () => {
    // Check if plugin is installed before enabling
    if (!config?.enabled && !pluginStatus?.installed) {
      return; // Can't enable without plugin
    }

    setSwitchLoading(true);
    if (config?.enabled) {
      await disableMTLS();
    } else {
      const success = await enableMTLS();
      if (success) {
        setShowSetupWarning(true);
      }
    }
    setSwitchLoading(false);
  };

  const handleSaveMiddlewareConfig = async () => {
    setSavingMiddleware(true);
    await updateMiddlewareConfig(middlewareForm);
    setSavingMiddleware(false);
  };

  const handleToggleTLSHardening = async () => {
    setTLSHardeningLoading(true);
    if (securityConfig?.tls_hardening_enabled) {
      await disableTLSHardening();
    } else {
      await enableTLSHardening();
    }
    setTLSHardeningLoading(false);
  };

  const handleToggleSecureHeaders = async () => {
    setSecureHeadersLoading(true);
    if (securityConfig?.secure_headers_enabled) {
      await disableSecureHeaders();
    } else {
      await enableSecureHeaders();
    }
    setSecureHeadersLoading(false);
  };

  const handleSaveSecureHeadersConfig = async () => {
    setSavingSecureHeaders(true);
    await updateSecureHeadersConfig(secureHeadersForm);
    setSavingSecureHeaders(false);
  };

  if (loading && !config) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Shield className="h-8 w-8" />
            Security
          </h1>
          <p className="text-muted-foreground">
            Manage mTLS (mutual TLS) for secure client authentication
          </p>
        </div>
        <div className="flex items-center gap-2">
          {/* Plugin Status Badge */}
          {pluginStatus?.installed ? (
            <Badge variant="outline" className="flex items-center gap-1">
              <CheckCircle className="h-4 w-4 text-green-500" />
              Plugin {pluginStatus.version || pluginStatus.recommended_version || 'installed'}
            </Badge>
          ) : (
            <Badge
              variant="outline"
              className="flex items-center gap-1 text-yellow-600"
            >
              <XCircle className="h-4 w-4" />
              Plugin Not Installed
            </Badge>
          )}
          {/* mTLS Status Badge */}
          {config?.enabled ? (
            <Badge variant="default" className="flex items-center gap-1">
              <ShieldCheck className="h-4 w-4" />
              mTLS Enabled
            </Badge>
          ) : (
            <Badge variant="secondary" className="flex items-center gap-1">
              <ShieldAlert className="h-4 w-4" />
              mTLS Disabled
            </Badge>
          )}
        </div>
      </div>

      {/* Error Alert */}
      {error && (
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription className="flex justify-between items-center">
            {error}
            <Button variant="outline" size="sm" onClick={clearError}>
              Dismiss
            </Button>
          </AlertDescription>
        </Alert>
      )}

      {/* Setup Warning */}
      {showSetupWarning && (
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>One-Time Plugin Setup Required</AlertTitle>
          <AlertDescription>
            <p className="mb-2">
              mTLS has been enabled. If this is your first time, complete the
              setup in the <strong>Setup</strong> tab, then restart Traefik.
            </p>
            <Button
              variant="outline"
              size="sm"
              className="mt-2"
              onClick={() => setShowSetupWarning(false)}
            >
              Got it
            </Button>
          </AlertDescription>
        </Alert>
      )}

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList className="flex-wrap h-auto gap-1">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="setup">Setup</TabsTrigger>
          <TabsTrigger value="ca">Certificate Authority</TabsTrigger>
          <TabsTrigger value="clients">Client Certificates</TabsTrigger>
          <TabsTrigger value="advanced">Advanced</TabsTrigger>
          <TabsTrigger value="tls-hardening">TLS Hardening</TabsTrigger>
          <TabsTrigger value="secure-headers">Secure Headers</TabsTrigger>
          <TabsTrigger value="guide">Import Guide</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            {/* mTLS Status Card */}
            <Card>
              <CardHeader>
                <CardTitle>mTLS Status</CardTitle>
                <CardDescription>
                  Enable or disable mutual TLS authentication globally
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="mtls-toggle">Enable mTLS</Label>
                    <p className="text-sm text-muted-foreground">
                      Require client certificates for all mTLS-enabled resources
                    </p>
                  </div>
                  <Switch
                    id="mtls-toggle"
                    checked={config?.enabled ?? false}
                    onCheckedChange={handleToggleMTLS}
                    disabled={
                      switchLoading ||
                      !config?.has_ca ||
                      !pluginStatus?.installed
                    }
                  />
                </div>
                {!pluginStatus?.installed && (
                  <Alert>
                    <Plug className="h-4 w-4" />
                    <AlertDescription>
                      Install the{" "}
                      <code className="bg-muted px-1 rounded">
                        mtlswhitelist
                      </code>{" "}
                      plugin first. See the <strong>Setup</strong> tab for
                      instructions.
                    </AlertDescription>
                  </Alert>
                )}
                {pluginStatus?.installed && !config?.has_ca && (
                  <Alert>
                    <Info className="h-4 w-4" />
                    <AlertDescription>
                      Create a Certificate Authority first to enable mTLS
                    </AlertDescription>
                  </Alert>
                )}
              </CardContent>
            </Card>

            {/* Statistics Card */}
            <Card>
              <CardHeader>
                <CardTitle>Certificate Statistics</CardTitle>
                <CardDescription>
                  Overview of your mTLS certificates
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">CA Status</p>
                    <p className="text-2xl font-bold">
                      {config?.has_ca ? "Configured" : "Not Set"}
                    </p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">
                      Client Certs
                    </p>
                    <p className="text-2xl font-bold">
                      {config?.client_count ?? 0}
                    </p>
                  </div>
                  {config?.ca_expiry && (
                    <div className="col-span-2 space-y-1">
                      <p className="text-sm text-muted-foreground">CA Expiry</p>
                      <p className="text-sm font-medium">
                        {new Date(config.ca_expiry).toLocaleDateString()}
                      </p>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* How it Works */}
          <Card>
            <CardHeader>
              <CardTitle>How mTLS Works</CardTitle>
              <CardDescription>
                Mutual TLS provides strong client authentication
              </CardDescription>
            </CardHeader>
            <CardContent className="prose prose-sm dark:prose-invert max-w-none">
              <p>
                With mTLS enabled, clients must present a valid certificate
                signed by your CA to access protected resources. This provides:
              </p>
              <ul>
                <li>
                  <strong>Strong Authentication:</strong> Only clients with
                  valid certificates can connect
                </li>
                <li>
                  <strong>No Login Page:</strong> Unauthorized users are
                  rejected at the TLS level
                </li>
                <li>
                  <strong>Per-Resource Control:</strong> Enable mTLS selectively
                  on specific routers
                </li>
              </ul>
              <p className="text-muted-foreground">
                After enabling mTLS globally, go to Resources and enable it on
                individual routers.
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Setup Tab */}
        <TabsContent value="setup" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>One-Time Traefik Setup</CardTitle>
              <CardDescription>
                Complete these steps once to enable mTLS authentication
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Step 1 */}
              <div className="space-y-3">
                <div className="flex items-center gap-2">
                  <Badge
                    variant="outline"
                    className="h-6 w-6 rounded-full p-0 flex items-center justify-center"
                  >
                    1
                  </Badge>
                  <h4 className="font-semibold">
                    Add the mtlswhitelist plugin to your Traefik static
                    configuration
                  </h4>
                </div>
                <p className="text-sm text-muted-foreground ml-8">
                  Add this to your{" "}
                  <code className="bg-muted px-1 rounded">
                    traefik_config.yml
                  </code>{" "}
                  (or static config file):
                </p>
                <pre className="bg-muted p-4 rounded text-xs overflow-x-auto ml-8">
                  {`experimental:
  plugins:
    mtlswhitelist:
      moduleName: github.com/smerschjohann/mtlswhitelist
      version: ${pluginStatus?.recommended_version || pluginStatus?.version || 'v0.0.4'}`}
                </pre>
              </div>

              {/* Step 2 */}
              <div className="space-y-3">
                <div className="flex items-center gap-2">
                  <Badge
                    variant="outline"
                    className="h-6 w-6 rounded-full p-0 flex items-center justify-center"
                  >
                    2
                  </Badge>
                  <h4 className="font-semibold">
                    Add the certificates volume mount
                  </h4>
                </div>
                <p className="text-sm text-muted-foreground ml-8">
                  Add this volume to your Traefik service in{" "}
                  <code className="bg-muted px-1 rounded">
                    docker-compose.yml
                  </code>
                  :
                </p>
                <pre className="bg-muted p-4 rounded text-xs overflow-x-auto ml-8">
                  {`volumes:
  - ./config/traefik/certs:/etc/traefik/certs:ro`}
                </pre>
              </div>

              {/* Step 3 */}
              <div className="space-y-3">
                <div className="flex items-center gap-2">
                  <Badge
                    variant="outline"
                    className="h-6 w-6 rounded-full p-0 flex items-center justify-center"
                  >
                    3
                  </Badge>
                  <h4 className="font-semibold">Restart Traefik</h4>
                </div>
                <p className="text-sm text-muted-foreground ml-8">
                  After making the above changes, restart Traefik to load the
                  plugin:
                </p>
                <pre className="bg-muted p-4 rounded text-xs overflow-x-auto ml-8">
                  {`docker compose down && docker compose up -d`}
                </pre>
              </div>

              <Alert className="mt-4">
                <Info className="h-4 w-4" />
                <AlertTitle>When to restart?</AlertTitle>
                <AlertDescription>
                  Only restart Traefik after completing steps 1 and 2. Once the
                  plugin is installed, you can enable/disable mTLS on individual
                  resources without restarting.
                </AlertDescription>
              </Alert>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>How It Works</CardTitle>
            </CardHeader>
            <CardContent className="prose prose-sm dark:prose-invert max-w-none">
              <p>
                The <code>mtlswhitelist</code> plugin validates client
                certificates at the middleware level, which means it works
                seamlessly with routers from any provider (HTTP API, Docker,
                File).
              </p>
              <p>
                When you enable mTLS on a resource, the <code>mtls-auth</code>{" "}
                middleware is automatically added to that router. The middleware
                checks:
              </p>
              <ul>
                <li>
                  That the client presented a certificate during TLS handshake
                </li>
                <li>That the certificate is signed by your CA</li>
                <li>That the certificate has not expired</li>
              </ul>
              <p className="text-muted-foreground">
                Requests without valid certificates receive a 403 Forbidden
                response.
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* CA Management Tab */}
        <TabsContent value="ca">
          <CAManager />
        </TabsContent>

        {/* Client Certificates Tab */}
        <TabsContent value="clients">
          <ClientCertList />
        </TabsContent>

        {/* Advanced Tab - Middleware Configuration */}
        <TabsContent value="advanced" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Middleware Configuration</CardTitle>
              <CardDescription>
                Configure advanced options for the mtlswhitelist plugin
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Reject Message */}
              <div className="space-y-2">
                <Label htmlFor="reject-message">Reject Message</Label>
                <p className="text-sm text-muted-foreground">
                  Message returned when certificate validation fails
                </p>
                <Input
                  id="reject-message"
                  value={middlewareForm.reject_message}
                  onChange={(e) =>
                    setMiddlewareForm((prev) => ({
                      ...prev,
                      reject_message: e.target.value,
                    }))
                  }
                  placeholder="Access denied: Valid client certificate required"
                />
              </div>

              {/* Refresh Interval */}
              <div className="space-y-2">
                <Label htmlFor="refresh-interval">Refresh Interval (seconds)</Label>
                <p className="text-sm text-muted-foreground">
                  How often to refresh external data sources (if configured)
                </p>
                <Input
                  id="refresh-interval"
                  type="number"
                  value={middlewareForm.refresh_interval}
                  onChange={(e) =>
                    setMiddlewareForm((prev) => ({
                      ...prev,
                      refresh_interval: parseInt(e.target.value) || 300,
                    }))
                  }
                  min={60}
                  max={3600}
                />
              </div>

              {/* Request Headers */}
              <div className="space-y-2">
                <Label htmlFor="request-headers">Request Headers (JSON)</Label>
                <p className="text-sm text-muted-foreground">
                  Headers to add to requests with certificate info. Example:{" "}
                  <code className="bg-muted px-1 rounded">
                    {`{"X-Client-CN": "{{ .Subject.CommonName }}"}`}
                  </code>
                </p>
                <Textarea
                  id="request-headers"
                  value={middlewareForm.request_headers}
                  onChange={(e) =>
                    setMiddlewareForm((prev) => ({
                      ...prev,
                      request_headers: e.target.value,
                    }))
                  }
                  placeholder='{"X-Client-CN": "{{ .Subject.CommonName }}"}'
                  rows={3}
                  className="font-mono text-sm"
                />
              </div>

              {/* Rules */}
              <div className="space-y-2">
                <Label htmlFor="rules">Validation Rules (JSON)</Label>
                <p className="text-sm text-muted-foreground">
                  Advanced rules for certificate validation. Leave empty for default
                  CA validation only.
                </p>
                <Textarea
                  id="rules"
                  value={middlewareForm.rules}
                  onChange={(e) =>
                    setMiddlewareForm((prev) => ({
                      ...prev,
                      rules: e.target.value,
                    }))
                  }
                  placeholder='[{"type": "AllOf", "rules": [{"type": "Header", "key": "Subject.CommonName", "value": "admin.*"}]}]'
                  rows={4}
                  className="font-mono text-sm"
                />
              </div>

              <Button
                onClick={handleSaveMiddlewareConfig}
                disabled={savingMiddleware}
              >
                {savingMiddleware ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Saving...
                  </>
                ) : (
                  "Save Configuration"
                )}
              </Button>
            </CardContent>
          </Card>

          {/* Documentation */}
          <Card>
            <CardHeader>
              <CardTitle>Rule Types</CardTitle>
              <CardDescription>
                Available rule types for the mtlswhitelist plugin
              </CardDescription>
            </CardHeader>
            <CardContent className="prose prose-sm dark:prose-invert max-w-none">
              <ul>
                <li>
                  <strong>AllOf:</strong> All nested rules must match
                </li>
                <li>
                  <strong>AnyOf:</strong> At least one nested rule must match
                </li>
                <li>
                  <strong>NoneOf:</strong> None of the nested rules must match
                </li>
                <li>
                  <strong>Header:</strong> Match certificate field (e.g.,{" "}
                  <code>Subject.CommonName</code>)
                </li>
                <li>
                  <strong>IPRange:</strong> Match client IP against CIDR range
                </li>
              </ul>
              <p className="text-muted-foreground">
                See the{" "}
                <a
                  href="https://github.com/smerschjohann/mtlswhitelist"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  plugin documentation
                </a>{" "}
                for more details.
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* TLS Hardening Tab */}
        <TabsContent value="tls-hardening" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Lock className="h-5 w-5" />
                TLS Hardening
              </CardTitle>
              <CardDescription>
                Configure TLS security settings for improved protection
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="tls-hardening-toggle">Enable TLS Hardening</Label>
                  <p className="text-sm text-muted-foreground">
                    Apply hardened TLS settings to resources (TLS 1.2+, secure ciphers)
                  </p>
                </div>
                <Switch
                  id="tls-hardening-toggle"
                  checked={securityConfig?.tls_hardening_enabled ?? false}
                  onCheckedChange={handleToggleTLSHardening}
                  disabled={tlsHardeningLoading}
                />
              </div>

              <Alert>
                <Info className="h-4 w-4" />
                <AlertDescription>
                  TLS Hardening is automatically disabled when mTLS is active on a resource.
                  mTLS already includes TLS hardening via the <code className="bg-muted px-1 rounded">mtls-verify</code> options.
                </AlertDescription>
              </Alert>

              {securityConfig?.tls_hardening_enabled && (
                <div className="space-y-4 pt-4 border-t">
                  <h4 className="font-medium">Applied TLS Settings</h4>
                  <div className="grid gap-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Minimum Version:</span>
                      <span className="font-mono">TLS 1.2</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Maximum Version:</span>
                      <span className="font-mono">TLS 1.3</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">SNI Strict:</span>
                      <span className="font-mono">true</span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Cipher Suites:</span>
                      <ul className="mt-1 ml-4 font-mono text-xs space-y-1">
                        <li>TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256</li>
                        <li>TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256</li>
                        <li>TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384</li>
                        <li>TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384</li>
                        <li>TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256</li>
                        <li>TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256</li>
                      </ul>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Curve Preferences:</span>
                      <span className="font-mono ml-2">X25519, CurveP384, CurveP521</span>
                    </div>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Per-Resource Configuration</CardTitle>
              <CardDescription>
                Enable TLS hardening on individual resources
              </CardDescription>
            </CardHeader>
            <CardContent className="prose prose-sm dark:prose-invert max-w-none">
              <p>
                After enabling TLS hardening globally, you can enable it on individual
                resources from the Resources page. Each resource will then use the
                <code className="bg-muted px-1 rounded">tls-hardened</code> TLS options.
              </p>
              <p className="text-muted-foreground">
                Note: The TLS hardening toggle is hidden for resources with mTLS enabled,
                as mTLS already provides equivalent security.
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Secure Headers Tab */}
        <TabsContent value="secure-headers" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileKey className="h-5 w-5" />
                Secure Headers
              </CardTitle>
              <CardDescription>
                Configure security response headers for protected resources
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="secure-headers-toggle">Enable Secure Headers</Label>
                  <p className="text-sm text-muted-foreground">
                    Add security headers to responses for enabled resources
                  </p>
                </div>
                <Switch
                  id="secure-headers-toggle"
                  checked={securityConfig?.secure_headers_enabled ?? false}
                  onCheckedChange={handleToggleSecureHeaders}
                  disabled={secureHeadersLoading}
                />
              </div>

              {securityConfig?.secure_headers_enabled && (
                <div className="space-y-4 pt-4 border-t">
                  <h4 className="font-medium">Header Configuration</h4>

                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="x-content-type-options">X-Content-Type-Options</Label>
                      <Input
                        id="x-content-type-options"
                        value={secureHeadersForm.x_content_type_options}
                        onChange={(e) =>
                          setSecureHeadersForm((prev) => ({
                            ...prev,
                            x_content_type_options: e.target.value,
                          }))
                        }
                        placeholder="nosniff"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="x-frame-options">X-Frame-Options</Label>
                      <Input
                        id="x-frame-options"
                        value={secureHeadersForm.x_frame_options}
                        onChange={(e) =>
                          setSecureHeadersForm((prev) => ({
                            ...prev,
                            x_frame_options: e.target.value,
                          }))
                        }
                        placeholder="SAMEORIGIN"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="x-xss-protection">X-XSS-Protection</Label>
                      <Input
                        id="x-xss-protection"
                        value={secureHeadersForm.x_xss_protection}
                        onChange={(e) =>
                          setSecureHeadersForm((prev) => ({
                            ...prev,
                            x_xss_protection: e.target.value,
                          }))
                        }
                        placeholder="1; mode=block"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="hsts">Strict-Transport-Security (HSTS)</Label>
                      <Input
                        id="hsts"
                        value={secureHeadersForm.hsts}
                        onChange={(e) =>
                          setSecureHeadersForm((prev) => ({
                            ...prev,
                            hsts: e.target.value,
                          }))
                        }
                        placeholder="max-age=31536000; includeSubDomains"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="referrer-policy">Referrer-Policy</Label>
                      <Input
                        id="referrer-policy"
                        value={secureHeadersForm.referrer_policy}
                        onChange={(e) =>
                          setSecureHeadersForm((prev) => ({
                            ...prev,
                            referrer_policy: e.target.value,
                          }))
                        }
                        placeholder="strict-origin-when-cross-origin"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="csp">Content-Security-Policy (optional)</Label>
                      <Textarea
                        id="csp"
                        value={secureHeadersForm.csp}
                        onChange={(e) =>
                          setSecureHeadersForm((prev) => ({
                            ...prev,
                            csp: e.target.value,
                          }))
                        }
                        placeholder="default-src 'self'; script-src 'self' 'unsafe-inline'"
                        rows={2}
                        className="font-mono text-sm"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="permissions-policy">Permissions-Policy (optional)</Label>
                      <Input
                        id="permissions-policy"
                        value={secureHeadersForm.permissions_policy}
                        onChange={(e) =>
                          setSecureHeadersForm((prev) => ({
                            ...prev,
                            permissions_policy: e.target.value,
                          }))
                        }
                        placeholder="geolocation=(), microphone=()"
                      />
                    </div>
                  </div>

                  <Button
                    onClick={handleSaveSecureHeadersConfig}
                    disabled={savingSecureHeaders}
                  >
                    {savingSecureHeaders ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Saving...
                      </>
                    ) : (
                      "Save Configuration"
                    )}
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Per-Resource Configuration</CardTitle>
              <CardDescription>
                Enable secure headers on individual resources
              </CardDescription>
            </CardHeader>
            <CardContent className="prose prose-sm dark:prose-invert max-w-none">
              <p>
                After enabling secure headers globally and configuring the header values,
                you can enable it on individual resources from the Resources page.
                Each enabled resource will have the configured security headers added
                to all responses.
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Import Guide Tab */}
        <TabsContent value="guide">
          <CertImportGuide />
        </TabsContent>
      </Tabs>

      {/* Security Error Alert */}
      {securityError && (
        <Alert variant="destructive" className="mt-4">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Security Error</AlertTitle>
          <AlertDescription className="flex justify-between items-center">
            {securityError}
            <Button variant="outline" size="sm" onClick={clearSecurityError}>
              Dismiss
            </Button>
          </AlertDescription>
        </Alert>
      )}
    </div>
  );
}
