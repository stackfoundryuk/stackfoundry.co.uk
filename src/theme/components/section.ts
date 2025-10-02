const Section = {
  baseStyle: {
    pt: 28,
    pb: 28,
    px: [4, null],
  },
  variants: {
    subtle: {},
    solid: {
      bg: "primary.400",
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    alternate: ({ colorMode }: any) => ({
      bg: colorMode === "dark" ? "gray.800" : "gray.50",
    }),
  },
  defaultProps: {
    variant: "subtle",
  },
};

export default Section;
