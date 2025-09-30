import { Footer } from "@/components/organisms/Footer";
import { Header } from "@/components/organisms/Header";
import { Providers } from "@/providers";

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Providers>
          <Header />
          <main>{children}</main>
          <Footer />
        </Providers>
      </body>
    </html>
  );
}
