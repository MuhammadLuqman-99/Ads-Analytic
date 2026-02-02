"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { formatDistanceToNow } from "date-fns";

interface UserProfile {
  userId: string;
  email: string;
  name: string;
  firstSeenAt: string;
  lastSeenAt: string;
  totalEvents: number;
  totalSessions: number;
}

interface UsersTableProps {
  title: string;
  users: UserProfile[];
  showChurnBadge?: boolean;
  className?: string;
}

export function UsersTable({ title, users, showChurnBadge = false, className }: UsersTableProps) {
  const formatDate = (dateStr: string) => {
    try {
      return formatDistanceToNow(new Date(dateStr), { addSuffix: true });
    } catch {
      return dateStr;
    }
  };

  return (
    <Card className={cn("bg-white", className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-semibold text-slate-900">{title}</CardTitle>
          <Badge variant="secondary">{users.length} users</Badge>
        </div>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>User</TableHead>
              <TableHead>First Seen</TableHead>
              <TableHead>Last Seen</TableHead>
              <TableHead className="text-right">Events</TableHead>
              <TableHead className="text-right">Sessions</TableHead>
              {showChurnBadge && <TableHead>Status</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={showChurnBadge ? 6 : 5} className="text-center py-8">
                  <span className="text-slate-500">No users found</span>
                </TableCell>
              </TableRow>
            ) : (
              users.map((user) => (
                <TableRow key={user.userId}>
                  <TableCell>
                    <div>
                      <div className="font-medium text-slate-900">
                        {user.name || "Unknown"}
                      </div>
                      <div className="text-sm text-slate-500">{user.email}</div>
                    </div>
                  </TableCell>
                  <TableCell className="text-sm text-slate-600">
                    {formatDate(user.firstSeenAt)}
                  </TableCell>
                  <TableCell className="text-sm text-slate-600">
                    {formatDate(user.lastSeenAt)}
                  </TableCell>
                  <TableCell className="text-right font-medium">
                    {user.totalEvents.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right font-medium">
                    {user.totalSessions.toLocaleString()}
                  </TableCell>
                  {showChurnBadge && (
                    <TableCell>
                      <Badge variant="destructive">Churned</Badge>
                    </TableCell>
                  )}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

export function UsersTableSkeleton() {
  return (
    <Card className="bg-white">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="h-6 w-32 bg-slate-200 rounded animate-pulse" />
          <div className="h-5 w-20 bg-slate-200 rounded animate-pulse" />
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="flex items-center gap-4 animate-pulse">
              <div className="flex-1">
                <div className="h-4 w-32 bg-slate-200 rounded" />
                <div className="mt-1 h-3 w-48 bg-slate-200 rounded" />
              </div>
              <div className="h-4 w-16 bg-slate-200 rounded" />
              <div className="h-4 w-16 bg-slate-200 rounded" />
              <div className="h-4 w-12 bg-slate-200 rounded" />
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
