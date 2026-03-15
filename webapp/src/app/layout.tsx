import type { Metadata, Viewport } from "next";
import { Rubik } from "next/font/google";
import { TelegramProvider } from "@/providers/TelegramProvider";
import { BottomNav } from "@/components/layout/BottomNav";
import { Header } from "@/components/layout/Header";
import Script from "next/script";
import "./globals.css";

const rubik = Rubik({ subsets: ["latin", "cyrillic"], variable: "--font-rubik" });

export const metadata: Metadata = {
  title: "ForgePath",
  description: "Language learning progress",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="dark">
      <head>
        <Script
          src="https://telegram.org/js/telegram-web-app.js"
          strategy="beforeInteractive"
        />
      </head>
      <body className={`${rubik.variable} font-sans antialiased`}>
        <TelegramProvider>
          <Header />
          <main className="px-4 pt-2 pb-24">{children}</main>
          <BottomNav />
        </TelegramProvider>
      </body>
    </html>
  );
}
