"use client";

import { Box, Flex, HStack, Link, Text } from "@chakra-ui/react";
import NextLink from "next/link";

export const Header = () => (
  <Box as="header" bg="darkGray.900" color="white" py={4} px={8}>
    <Flex maxW="1200px" mx="auto" align="center" justify="space-between">
      <NextLink href="/" passHref>
        <Text fontSize="xl" fontWeight="bold" cursor="pointer">
          StackFoundry
        </Text>
      </NextLink>

      <HStack spacing={6}>
        <NextLink href="#services" passHref>
          <Link>Services</Link>
        </NextLink>
        <NextLink href="#about" passHref>
          <Link>About</Link>
        </NextLink>
        <NextLink href="#contact" passHref>
          <Link>Contact</Link>
        </NextLink>
      </HStack>
    </Flex>
  </Box>
);
