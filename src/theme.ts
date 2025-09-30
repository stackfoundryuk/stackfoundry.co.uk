import { extendTheme, ThemeConfig } from "@chakra-ui/react";

const config: ThemeConfig = {
  initialColorMode: "light",
  useSystemColorMode: false,
};

const colors = {
  brand: {
    50: "#fff5f6",
    100: "#ffe6e8",
    200: "#ffbfc3",
    300: "#ff9aa0",
    400: "#ff6a6f",
    500: "#c96b28ff", // your main orange
    600: "#c65257",
    700: "#9a3d3f",
    800: "#6a2a29",
    900: "#3b1514",
  },
  darkGray: {
    50: "#f7f7f7",
    100: "#efefef",
    200: "#e0dfdf",
    300: "#cfcfcf",
    400: "#bfbfbf",
    500: "#9e9c9c",
    600: "#777474",
    700: "#5a5858",
    800: "#3f3d3d",
    900: "#393636", // your main dark gray
  },
  accent: {
    500: "#2bb6a3",
  },
  secondary: {
    500: "#6b5bd3",
  },
};

const styles = {
  global: {
    body: {
      bg: "white",
      color: "darkGray.900",
    },
    a: {
      color: "brand.500",
      _hover: { color: "brand.600", textDecoration: "none" },
    },
  },
};

const components = {
  Button: {
    baseStyle: {
      borderRadius: "8px",
      _focus: { boxShadow: "outline" },
    },
    variants: {
      brandSolid: {
        bg: "brand.500",
        color: "white",
        _hover: {
          bg: "white",
          color: "brand.700",
          border: "1px solid",
          borderColor: "brand.700",
          transform: "translateY(-1px)",
        },
        _active: { transform: "translateY(0)" },
      },
    },
  },
};

const theme = extendTheme({
  config,
  colors,
  styles,
  components,
  fonts: {
    heading: "Inter, sans-serif",
    body: "Inter, sans-serif",
  },
});

export default theme;
