import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Chrome, Smartphone, Monitor, Apple, Info, Shield } from 'lucide-react'

export function CertImportGuide() {
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Certificate Import Guide
          </CardTitle>
          <CardDescription>
            Instructions for importing client certificates on various platforms
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Alert className="mb-6">
            <Info className="h-4 w-4" />
            <AlertTitle>Before You Start</AlertTitle>
            <AlertDescription>
              <ul className="list-disc list-inside mt-2 space-y-1">
                <li>Download the .p12 file from the Client Certificates tab</li>
                <li>Keep your certificate password ready</li>
                <li>Transfer the file securely to your device (avoid email)</li>
              </ul>
            </AlertDescription>
          </Alert>

          <Accordion type="single" collapsible className="w-full">
            {/* Firefox */}
            <AccordionItem value="firefox">
              <AccordionTrigger>
                <div className="flex items-center gap-2">
                  <Monitor className="h-4 w-4" />
                  Firefox
                  <Badge variant="secondary">Desktop</Badge>
                </div>
              </AccordionTrigger>
              <AccordionContent>
                <ol className="list-decimal list-inside space-y-2 text-sm">
                  <li>Open Firefox and go to <strong>Settings</strong></li>
                  <li>Navigate to <strong>Privacy & Security</strong></li>
                  <li>Scroll down to <strong>Certificates</strong> section</li>
                  <li>Click <strong>View Certificates</strong></li>
                  <li>Select the <strong>Your Certificates</strong> tab</li>
                  <li>Click <strong>Import...</strong></li>
                  <li>Select your .p12 file</li>
                  <li>Enter the certificate password when prompted</li>
                  <li>The certificate should now appear in the list</li>
                </ol>
                <p className="mt-4 text-muted-foreground">
                  When visiting a protected site, Firefox will prompt you to select a certificate.
                </p>
              </AccordionContent>
            </AccordionItem>

            {/* Chrome / Edge */}
            <AccordionItem value="chrome">
              <AccordionTrigger>
                <div className="flex items-center gap-2">
                  <Chrome className="h-4 w-4" />
                  Chrome / Edge
                  <Badge variant="secondary">Desktop</Badge>
                </div>
              </AccordionTrigger>
              <AccordionContent>
                <ol className="list-decimal list-inside space-y-2 text-sm">
                  <li>Open Chrome/Edge and go to <strong>Settings</strong></li>
                  <li>Search for "certificates" or navigate to <strong>Privacy and security</strong></li>
                  <li>Click <strong>Security</strong></li>
                  <li>Click <strong>Manage certificates</strong></li>
                  <li>In the Certificate Manager, select the <strong>Personal</strong> tab</li>
                  <li>Click <strong>Import...</strong></li>
                  <li>Follow the Certificate Import Wizard</li>
                  <li>Select your .p12 file</li>
                  <li>Enter the certificate password</li>
                  <li>Complete the wizard</li>
                </ol>
                <p className="mt-4 text-muted-foreground">
                  Note: On Windows, Chrome and Edge use the Windows certificate store.
                  On macOS, they use the Keychain.
                </p>
              </AccordionContent>
            </AccordionItem>

            {/* iOS */}
            <AccordionItem value="ios">
              <AccordionTrigger>
                <div className="flex items-center gap-2">
                  <Apple className="h-4 w-4" />
                  iOS (iPhone/iPad)
                  <Badge variant="secondary">Mobile</Badge>
                </div>
              </AccordionTrigger>
              <AccordionContent>
                <ol className="list-decimal list-inside space-y-2 text-sm">
                  <li>Transfer the .p12 file to your device via AirDrop, iCloud, or a secure method</li>
                  <li>Tap on the .p12 file to open it</li>
                  <li>You'll see a message: <strong>"Profile Downloaded"</strong></li>
                  <li>Go to <strong>Settings</strong></li>
                  <li>You'll see <strong>"Profile Downloaded"</strong> near the top - tap it</li>
                  <li>Tap <strong>Install</strong> in the top right</li>
                  <li>Enter your device passcode if prompted</li>
                  <li>Enter the certificate password</li>
                  <li>Tap <strong>Install</strong> again to confirm</li>
                  <li>The certificate is now installed</li>
                </ol>
                <Alert className="mt-4">
                  <Info className="h-4 w-4" />
                  <AlertDescription>
                    When using Safari to visit a protected site, iOS will automatically
                    present your certificate for authentication.
                  </AlertDescription>
                </Alert>
              </AccordionContent>
            </AccordionItem>

            {/* Android */}
            <AccordionItem value="android">
              <AccordionTrigger>
                <div className="flex items-center gap-2">
                  <Smartphone className="h-4 w-4" />
                  Android
                  <Badge variant="secondary">Mobile</Badge>
                </div>
              </AccordionTrigger>
              <AccordionContent>
                <ol className="list-decimal list-inside space-y-2 text-sm">
                  <li>Transfer the .p12 file to your device</li>
                  <li>Go to <strong>Settings</strong></li>
                  <li>Navigate to <strong>Security</strong> (or <strong>Biometrics and security</strong>)</li>
                  <li>Find <strong>Encryption & credentials</strong></li>
                  <li>Tap <strong>Install a certificate</strong></li>
                  <li>Select <strong>VPN & app user certificate</strong></li>
                  <li>Browse to and select your .p12 file</li>
                  <li>Enter the certificate password</li>
                  <li>Give the certificate a name (optional)</li>
                  <li>Tap <strong>OK</strong> to install</li>
                </ol>
                <p className="mt-4 text-muted-foreground">
                  Note: The exact menu path may vary depending on your Android version
                  and device manufacturer.
                </p>
              </AccordionContent>
            </AccordionItem>

            {/* macOS Keychain */}
            <AccordionItem value="macos">
              <AccordionTrigger>
                <div className="flex items-center gap-2">
                  <Apple className="h-4 w-4" />
                  macOS (Keychain)
                  <Badge variant="secondary">Desktop</Badge>
                </div>
              </AccordionTrigger>
              <AccordionContent>
                <ol className="list-decimal list-inside space-y-2 text-sm">
                  <li>Double-click the .p12 file</li>
                  <li>Keychain Access will open</li>
                  <li>Select the keychain to add the certificate to (usually "login")</li>
                  <li>Enter the certificate password</li>
                  <li>If prompted, enter your macOS password to allow access</li>
                  <li>The certificate and private key are now in your keychain</li>
                </ol>
                <p className="mt-4 text-muted-foreground">
                  Safari and other macOS apps will automatically use certificates from the Keychain.
                </p>
              </AccordionContent>
            </AccordionItem>

            {/* Windows */}
            <AccordionItem value="windows">
              <AccordionTrigger>
                <div className="flex items-center gap-2">
                  <Monitor className="h-4 w-4" />
                  Windows
                  <Badge variant="secondary">Desktop</Badge>
                </div>
              </AccordionTrigger>
              <AccordionContent>
                <ol className="list-decimal list-inside space-y-2 text-sm">
                  <li>Double-click the .p12 file</li>
                  <li>The Certificate Import Wizard will open</li>
                  <li>Select <strong>Current User</strong> and click Next</li>
                  <li>Confirm the file path and click Next</li>
                  <li>Enter the certificate password</li>
                  <li>Check <strong>Mark this key as exportable</strong> if you want to back it up later</li>
                  <li>Click Next</li>
                  <li>Select <strong>Automatically select the certificate store</strong></li>
                  <li>Click Next, then Finish</li>
                </ol>
                <p className="mt-4 text-muted-foreground">
                  Chrome and Edge will automatically use certificates from the Windows certificate store.
                </p>
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </CardContent>
      </Card>

      {/* Security Best Practices */}
      <Card>
        <CardHeader>
          <CardTitle>Security Best Practices</CardTitle>
          <CardDescription>
            Tips for keeping your certificates secure
          </CardDescription>
        </CardHeader>
        <CardContent className="prose prose-sm dark:prose-invert max-w-none">
          <ul>
            <li>
              <strong>Transfer certificates securely:</strong> Never send .p12 files via unencrypted
              email. Use secure channels like AirDrop, Signal, or hand-deliver on a USB drive.
            </li>
            <li>
              <strong>Use strong passwords:</strong> The .p12 password protects your private key.
              Use a unique, strong password for each certificate.
            </li>
            <li>
              <strong>Backup your CA:</strong> The CA private key cannot be recovered if lost.
              Keep a secure backup of your CA certificates.
            </li>
            <li>
              <strong>Revoke compromised certificates:</strong> If a device is lost or stolen,
              immediately revoke its certificate in the Client Certificates tab.
            </li>
            <li>
              <strong>Monitor certificate expiry:</strong> Set reminders to renew certificates
              before they expire to avoid service disruption.
            </li>
          </ul>
        </CardContent>
      </Card>
    </div>
  )
}
