"use client"

import { History, Images, MenuIcon, PenSquare } from "lucide-react"
import Link from "next/link" // <--- 1. Import the Link component
import { usePathname } from "next/navigation"
import { useEffect } from "react"
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarTrigger,
    useSidebar,
} from "./ui/sidebar"
import { cn } from "~/lib/utils"

const items = [
    { title: "New", url: "/", icon: PenSquare },
    { title: "History", url: "/history", icon: History },
    { title: "Gallery", url: "/gallery", icon: Images },
]

export function AppSidebar() {
    const path = usePathname()
    const {
        state,
        open,
        setOpen,
        openMobile,
        setOpenMobile,
        isMobile,
    } = useSidebar()

    const isExpanded = open || openMobile || state === "expanded"

    useEffect(() => {
        if (isMobile) {
            setOpen(openMobile)
        } else {
            setOpenMobile(open)
        }
    }, [isMobile, open, openMobile, setOpen, setOpenMobile])

    useEffect(() => {
        if (isMobile) {
            setOpenMobile(false)
        }
    }, [path])

    return (
        <>
            {/* Mobile menu trigger */}
            {isMobile && (
                <div className="flex justify-center ml-1 mt-2">
                    <SidebarTrigger icon={MenuIcon} />
                </div>
            )}

            <Sidebar collapsible="icon">
                <SidebarHeader>
                    <div
                        className={cn(
                            "flex items-center",
                            isExpanded ? "justify-between ml-1" : "justify-center"
                        )}
                    >
                        {isExpanded && <span className="font-bold">Selfier</span>}
                        <SidebarTrigger />
                    </div>
                </SidebarHeader>

                <SidebarContent>
                    <SidebarGroup>
                        <SidebarGroupContent>
                            <SidebarMenu>
                                {items.map((item) => (
                                    <SidebarMenuItem key={item.title}>
                                        <SidebarMenuButton asChild isActive={path === item.url}>
                                            <Link href={item.url} className="flex items-center gap-2">
                                                <item.icon className="w-4 h-4" />
                                                <span>{item.title}</span>
                                            </Link>
                                        </SidebarMenuButton>
                                    </SidebarMenuItem>
                                ))}
                            </SidebarMenu>
                        </SidebarGroupContent>
                    </SidebarGroup>
                </SidebarContent>

                <SidebarFooter />
            </Sidebar>
        </>
    )
}
