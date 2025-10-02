"use client";

import { AuthProvider } from "@saas-ui/auth";
import { SaasProvider } from "@saas-ui/react";

import { theme } from "@/theme/index";

export function Provider(props: { children: React.ReactNode }) {
  return (
    <SaasProvider theme={theme}>
      <AuthProvider>{props.children}</AuthProvider>
    </SaasProvider>
  );
}
