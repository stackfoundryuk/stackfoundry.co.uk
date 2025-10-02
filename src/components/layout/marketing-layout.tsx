"use client";

import { Box, SkipNavContent, SkipNavLink } from "@chakra-ui/react";
import { ReactNode } from "react";

import { Footer, FooterProps } from "./footer";
import { Header, HeaderProps } from "./header";
import {
  AnnouncementBanner,
  AnnouncementBannerProps,
} from "../molecules/announcement-banner";

interface LayoutProps {
  children: ReactNode
  announcementProps?: AnnouncementBannerProps
  headerProps?: HeaderProps
  footerProps?: FooterProps
}

export const MarketingLayout: React.FC<LayoutProps> = (props) => {
  const { children, announcementProps, headerProps, footerProps } = props;
  return (
    <Box>
      <SkipNavLink>Skip to content</SkipNavLink>
      {announcementProps ? <AnnouncementBanner {...announcementProps} /> : null}
      <Header {...headerProps} />
      <Box as="main">
        <SkipNavContent />
        {children}
      </Box>
      <Footer {...footerProps} />
    </Box>
  );
};
