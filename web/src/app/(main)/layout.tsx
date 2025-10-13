import "~/styles/globals.css";

import { type Metadata } from "next";
import { Geist } from "next/font/google";
import { SidebarInset, SidebarProvider } from "~/components/ui/sidebar";
import { AppSidebar } from "~/components/app-sidebar";
import { ThemeProvider } from "~/components/theme-provider";
import { Toaster } from 'react-hot-toast';

export const metadata: Metadata = {
  title: "Selfier - Create Stunning Photos With Your Selfie",
  description: "Selfier - Create Stunning Photos With Your Selfie",
  icons: [{ rel: "icon", url: "/favicon.ico" }],
};

const geist = Geist({
  subsets: ["latin"],
  variable: "--font-geist-sans",
});

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {

  return (
    <html lang="en" suppressHydrationWarning className={`${geist.variable}`}>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Public+Sans:wght@100;200;300;400;500;600;700;800;900&display=swap" rel="stylesheet" />
        <link href="https://fonts.googleapis.com/css2?family=Inter:wght@100;200;300;400;500;600;700;800;900&display=swap" rel="stylesheet" />
      </head>
      <body suppressHydrationWarning className="flex min-h-svh flex-col items-center justify-center">
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem={true}
          disableTransitionOnChange
        ></ThemeProvider>

        <Toaster
          toastOptions={{
            className: "text-body border",
            style: {
              background: "var(--background-light)",
              color: "var(--foreground)",
            },
            success: {
              iconTheme: {
                primary: 'var(--color-success-300)',
                secondary: 'var(--color-success-800)',
              },
            },
            error: {
              iconTheme: {
                primary: 'var(--color-error-300)',
                secondary: 'var(--color-error-800)',
              },
            }
          }}
          position="bottom-right"
        />
        <SidebarProvider defaultOpen={false}>
          <AppSidebar />
          <SidebarInset>
            <main className="flex-1 overflow-y-auto">
              {children}
            </main>

          </SidebarInset>
        </SidebarProvider>
      </body>
    </html>
  );
}
