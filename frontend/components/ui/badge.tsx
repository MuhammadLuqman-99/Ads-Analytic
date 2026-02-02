import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
    "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2",
    {
        variants: {
            variant: {
                default:
                    "bg-indigo-500/20 text-indigo-400 border border-indigo-500/30",
                secondary:
                    "bg-slate-500/20 text-slate-300 border border-slate-500/30",
                success:
                    "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30",
                warning:
                    "bg-amber-500/20 text-amber-400 border border-amber-500/30",
                destructive:
                    "bg-red-500/20 text-red-400 border border-red-500/30",
                outline:
                    "bg-transparent text-slate-700 border border-slate-200",
                meta:
                    "bg-blue-500/20 text-blue-400 border border-blue-500/30",
                google:
                    "bg-red-500/20 text-red-400 border border-red-500/30",
                tiktok:
                    "bg-slate-500/20 text-slate-300 border border-slate-500/30",
                shopee:
                    "bg-orange-500/20 text-orange-400 border border-orange-500/30",
                linkedin:
                    "bg-blue-600/20 text-blue-500 border border-blue-600/30",
            },
        },
        defaultVariants: {
            variant: "default",
        },
    }
)

export interface BadgeProps
    extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> { }

function Badge({ className, variant, ...props }: BadgeProps) {
    return (
        <div className={cn(badgeVariants({ variant }), className)} {...props} />
    )
}

export { Badge, badgeVariants }
