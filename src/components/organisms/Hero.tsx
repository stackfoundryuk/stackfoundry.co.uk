"use client";

import { Button, Heading, Text } from "@chakra-ui/react";

import { MotionBox } from "@/motion";

export const Hero = () => (
  <MotionBox
    initial={{ opacity: 0, y: -20 }}
    animate={{ opacity: 1, y: 0 }}
    textAlign="center"
    py={32}
    px={4}
  >
    <Heading fontSize={["3xl","4xl","5xl"]} mb={6}>
      We Build Exceptional Software
    </Heading>
    <Text fontSize={["md","lg","xl"]} mb={8} maxW="700px" mx="auto">
      StackFoundry delivers high-quality, scalable software solutions. Full-stack, frontend, backend — we do it all.
    </Text>
    <Button
      colorScheme="brand"
      size="lg"
      _hover={{ bg: "brand.700" }}
    >
      Get in Touch
    </Button>
  </MotionBox>
);
