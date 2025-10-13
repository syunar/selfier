"use client";

import React from "react";
import { cn } from "~/lib/utils";

interface BadgeRootProps extends React.HTMLAttributes<HTMLDivElement> {
    variant?: "brand" | "neutral" | "error" | "warning" | "success";
    icon?: React.ReactNode;
    children?: React.ReactNode;
    iconRight?: React.ReactNode;
    className?: string;
}

const BadgeRoot = React.forwardRef<HTMLDivElement, BadgeRootProps>(
    (
        {
            variant = "brand",
            icon = null,
            children,
            iconRight = null,
            className,
            ...otherProps
        },
        ref
    ) => {
        return (
            <div
                ref={ref}
                {...otherProps}
                className={cn(
                    "group/97bdb082 flex h-6 items-center gap-1 rounded-md border border-solid border-brand-100 bg-brand-100 px-2",
                    {
                        "border-success-100 bg-success-100": variant === "success",
                        "border-warning-100 bg-warning-100": variant === "warning",
                        "border-error-100 bg-error-100": variant === "error",
                        "border-neutral-100 bg-neutral-100": variant === "neutral",
                    },
                    className
                )}
            >
                {icon && (
                    <span
                        className={cn(
                            "text-caption font-caption",
                            {
                                "text-success-800": variant === "success",
                                "text-warning-800": variant === "warning",
                                "text-error-700": variant === "error",
                                "text-neutral-700": variant === "neutral",
                            }
                        )}
                    >
                        {icon}
                    </span>
                )}

                {children && (
                    <span
                        className={cn(
                            "whitespace-nowrap text-caption font-caption",
                            {
                                "text-success-800": variant === "success",
                                "text-warning-800": variant === "warning",
                                "text-error-800": variant === "error",
                                "text-neutral-700": variant === "neutral",
                            }
                        )}
                    >
                        {children}
                    </span>
                )}

                {iconRight && (
                    <span
                        className={cn(
                            "text-caption font-caption",
                            {
                                "text-success-800": variant === "success",
                                "text-warning-800": variant === "warning",
                                "text-error-700": variant === "error",
                                "text-neutral-700": variant === "neutral",
                            }
                        )}
                    >
                        {iconRight}
                    </span>
                )}
            </div>
        );
    }
);

BadgeRoot.displayName = "BadgeRoot";

export const Badge = BadgeRoot;
