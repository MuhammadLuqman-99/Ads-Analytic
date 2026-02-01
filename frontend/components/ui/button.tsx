"use client"

import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
    "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-lg text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
    {
        variants: {
            variant: {
                default:
                    "bg-indigo-600 text-white shadow-lg shadow-indigo-500/25 hover:bg-indigo-500 focus-visible:ring-indigo-500",
                destructive:
                    "bg-red-600 text-white shadow-lg shadow-red-500/25 hover:bg-red-500 focus-visible:ring-red-500",
                outline:
                    "border border-slate-700 bg-transparent text-slate-300 hover:bg-slate-800 hover:text-white focus-visible:ring-slate-500",
                secondary:
                    "bg-slate-700 text-white shadow-lg hover:bg-slate-600 focus-visible:ring-slate-500",
                ghost:
                    "text-slate-400 hover:bg-slate-800 hover:text-white focus-visible:ring-slate-500",
                link: "text-indigo-400 underline-offset-4 hover:underline focus-visible:ring-indigo-500",
            },
            size: {
                default: "h-10 px-4 py-2",
                sm: "h-8 px-3 text-xs",
                lg: "h-12 px-6",
                icon: "h-10 w-10",
            },
        },
        defaultVariants: {
            variant: "default",
            size: "default",
        },
    }
)

export interface ButtonProps
    extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
    asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
    ({ className, variant, size, asChild = false, ...props }, ref) => {
        const Comp = asChild ? Slot : "button"
        return (
            <Comp
                className={cn(buttonVariants({ variant, size, className }))}
                ref={ref}
                {...props}
            />
        )
    }
)
Button.displayName = "Button"

export { Button, buttonVariants }
