export interface ITheme {
  name: string;
  // When true, Run:ai-specific UI elements are hidden (breadcrumb icon, tenant name, certain menu items and features).
  isWhiteLabel: boolean;
  color: {
    primaryColor: string;
    secondaryColor: string;
    secondaryColorOpacity: string;
    menuColor: string;
    menuSeparatorColor: string;
    menuSeparatorOpacity: string;
  };
  favicon: string;
  logo: {
    main: {
      light: { src: string; height: string }; // for dark backgrounds (sidebar corner)
      dark: { src: string; height: string }; // for light backgrounds (full-width page headers)
    };
    // Shown in the expanded side nav header and next to breadcrumbs in the page header.
    // When not provided: the expanded side nav still reserves header-height space at the top;
    // the page header shows nothing (no reserved space — breadcrumb starts from the left edge).
    sub?: string;
  };
}
