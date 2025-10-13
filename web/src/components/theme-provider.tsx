"use client"

import * as React from "react"
import { ThemeProvider as NextThemesProvider, useTheme } from "next-themes"

export function ThemeProvider({
    children,
    ...props
}: React.ComponentProps<typeof NextThemesProvider>) {
    const { setTheme } = useTheme()
    return <NextThemesProvider {...props}>
        <SetLightThemeOnce />
        {children}
    </NextThemesProvider>
}

function SetLightThemeOnce() {
    const { setTheme } = useTheme();

    React.useEffect(() => {
        setTheme("dark");
    }, [setTheme]);

    return null;
}
