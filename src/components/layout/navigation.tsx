/* eslint-disable import/order */
import { HStack, useDisclosure, useUpdateEffect } from "@chakra-ui/react";
import { usePathname } from "next/navigation";
import * as React from "react";

import { MobileNavButton, MobileNavContent } from "@/components/molecules/mobile-nav";
import { NavLink } from "@/components/molecules/nav-link";
import { useScrollSpy } from "@/hooks/use-scrollspy";

import siteConfig from "~/data/config";
import ThemeToggle from "./theme-toggle";



const Navigation: React.FC = () => {
  const mobileNav = useDisclosure();
  const path = usePathname();
  const activeId = useScrollSpy(
    siteConfig.header.links
      .filter(({ id }) => id)
      .map(({ id }) => `[id="${id}"]`),
    {
      threshold: 0.75,
    },
  );

  const mobileNavBtnRef = React.useRef<HTMLButtonElement>(null);

  useUpdateEffect(() => {
    mobileNavBtnRef.current?.focus();
  }, [mobileNav.isOpen]);

  return (
    <HStack spacing="2" flexShrink={0}>
      {siteConfig.header.links.map(({ href, id, ...props }, i) => {
        return (
          <NavLink
            display={["none", null, "block"]}
            href={href || `/#${id}`}
            key={i}
            isActive={
              !!(
                (id && activeId === id) ||
                (href && !!path?.match(new RegExp(href)))
              )
            }
            {...props}
          >
            {props.label}
          </NavLink>
        );
      })}

      <ThemeToggle />

      <MobileNavButton
        ref={mobileNavBtnRef}
        aria-label="Open Menu"
        onClick={mobileNav.onOpen}
      />

      <MobileNavContent isOpen={mobileNav.isOpen} onClose={mobileNav.onClose} />
    </HStack>
  );
};

export default Navigation;
