// src/providers.tsx (or wherever your ChakraProvider is)

"use client";

import "@fontsource/inter/400.css"; // Regular
import "@fontsource/inter/500.css"; // Medium
import "@fontsource/inter/700.css"; // Bold

import { ChakraProvider } from "@chakra-ui/react";

import theme from "./theme";

export const Providers = ({ children }: { children: React.ReactNode }) => {
  return <ChakraProvider theme={theme}>{children}</ChakraProvider>;
};
