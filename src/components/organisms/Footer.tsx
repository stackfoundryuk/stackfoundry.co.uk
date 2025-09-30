// app/components/Footer.tsx
"use client";

import { Box, Link, Text } from "@chakra-ui/react";

export const Footer = () => (
  <Box as="footer" bg="darkGray.900" color="white" py={6} px={8}>
    <Text textAlign="center" fontSize="sm">
      © {new Date().getFullYear()} StackFoundry. All rights reserved.
    </Text>
    <Text textAlign="center" fontSize="sm">
      <Link href="mailto:hello@stackfoundry.co.uk" color="brand.500">
        hello@stackfoundry.co.uk
      </Link>
    </Text>
  </Box>
);
