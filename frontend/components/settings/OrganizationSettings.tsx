"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Building2,
  Upload,
  UserPlus,
  MoreVertical,
  Mail,
  Shield,
  Trash2,
  X,
  Check,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";

type Role = "admin" | "member" | "viewer";

interface TeamMember {
  id: string;
  name: string;
  email: string;
  role: Role;
  avatar?: string;
  status: "active" | "pending";
  joinedAt: Date;
}

const mockTeamMembers: TeamMember[] = [
  {
    id: "1",
    name: "John Doe",
    email: "john@example.com",
    role: "admin",
    status: "active",
    joinedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 90),
  },
  {
    id: "2",
    name: "Jane Smith",
    email: "jane@example.com",
    role: "member",
    status: "active",
    joinedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 30),
  },
  {
    id: "3",
    name: "Bob Wilson",
    email: "bob@example.com",
    role: "viewer",
    status: "pending",
    joinedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 2),
  },
];

const roleConfig: Record<Role, { label: string; color: string; description: string }> = {
  admin: {
    label: "Admin",
    color: "bg-purple-100 text-purple-700",
    description: "Full access to all settings and data",
  },
  member: {
    label: "Member",
    color: "bg-blue-100 text-blue-700",
    description: "Can manage campaigns and view analytics",
  },
  viewer: {
    label: "Viewer",
    color: "bg-slate-100 text-slate-700",
    description: "Read-only access to analytics",
  },
};

const orgSchema = z.object({
  name: z.string().min(2, "Organization name is required"),
});

type OrgFormData = z.infer<typeof orgSchema>;

const inviteSchema = z.object({
  email: z.string().email("Please enter a valid email address"),
  role: z.enum(["admin", "member", "viewer"]),
});

type InviteFormData = z.infer<typeof inviteSchema>;

export function OrganizationSettings() {
  const [teamMembers, setTeamMembers] = useState<TeamMember[]>(mockTeamMembers);
  const [isUpdating, setIsUpdating] = useState(false);
  const [showInviteForm, setShowInviteForm] = useState(false);
  const [isInviting, setIsInviting] = useState(false);
  const [logoPreview, setLogoPreview] = useState<string | null>(null);

  const orgForm = useForm<OrgFormData>({
    resolver: zodResolver(orgSchema),
    defaultValues: {
      name: "My Organization",
    },
  });

  const inviteForm = useForm<InviteFormData>({
    resolver: zodResolver(inviteSchema),
    defaultValues: {
      email: "",
      role: "member",
    },
  });

  const onOrgSubmit = async (data: OrgFormData) => {
    setIsUpdating(true);
    await new Promise((resolve) => setTimeout(resolve, 1000));
    console.log("Organization updated:", data);
    setIsUpdating(false);
  };

  const onInviteSubmit = async (data: InviteFormData) => {
    setIsInviting(true);
    await new Promise((resolve) => setTimeout(resolve, 1000));

    const newMember: TeamMember = {
      id: String(Date.now()),
      name: data.email.split("@")[0],
      email: data.email,
      role: data.role,
      status: "pending",
      joinedAt: new Date(),
    };

    setTeamMembers((prev) => [...prev, newMember]);
    inviteForm.reset();
    setShowInviteForm(false);
    setIsInviting(false);
  };

  const handleLogoChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onloadend = () => {
        setLogoPreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleRemoveMember = (memberId: string) => {
    if (confirm("Are you sure you want to remove this team member?")) {
      setTeamMembers((prev) => prev.filter((m) => m.id !== memberId));
    }
  };

  const handleChangeRole = (memberId: string, newRole: Role) => {
    setTeamMembers((prev) =>
      prev.map((m) => (m.id === memberId ? { ...m, role: newRole } : m))
    );
  };

  const currentUser = teamMembers.find((m) => m.id === "1");

  return (
    <div className="space-y-6 max-w-3xl">
      {/* Organization Details */}
      <Card className="bg-white border-slate-200">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-slate-900">
            <Building2 className="h-5 w-5" />
            Organization Details
          </CardTitle>
          <CardDescription>
            Manage your organization&apos;s profile and branding
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={orgForm.handleSubmit(onOrgSubmit)} className="space-y-6">
            {/* Logo Upload */}
            <div className="space-y-2">
              <Label>Organization Logo</Label>
              <div className="flex items-center gap-4">
                <div className="w-20 h-20 rounded-lg bg-slate-100 border-2 border-dashed border-slate-300 flex items-center justify-center overflow-hidden">
                  {logoPreview ? (
                    <img
                      src={logoPreview}
                      alt="Logo preview"
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <Building2 className="h-8 w-8 text-slate-400" />
                  )}
                </div>
                <div>
                  <label htmlFor="logo-upload">
                    <Button type="button" variant="outline" asChild>
                      <span>
                        <Upload className="h-4 w-4 mr-2" />
                        Upload Logo
                      </span>
                    </Button>
                  </label>
                  <input
                    id="logo-upload"
                    type="file"
                    accept="image/*"
                    className="hidden"
                    onChange={handleLogoChange}
                  />
                  <p className="text-xs text-slate-500 mt-1">
                    PNG, JPG up to 2MB. Recommended 200x200px.
                  </p>
                </div>
              </div>
            </div>

            {/* Organization Name */}
            <div className="space-y-2">
              <Label htmlFor="orgName">Organization Name</Label>
              <Input
                id="orgName"
                {...orgForm.register("name")}
                placeholder="Enter organization name"
              />
              {orgForm.formState.errors.name && (
                <p className="text-sm text-red-600">
                  {orgForm.formState.errors.name.message}
                </p>
              )}
            </div>

            <div className="flex justify-end">
              <Button type="submit" disabled={isUpdating}>
                {isUpdating ? "Saving..." : "Save Changes"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {/* Team Members */}
      <Card className="bg-white border-slate-200">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2 text-slate-900">
                <Shield className="h-5 w-5" />
                Team Members
              </CardTitle>
              <CardDescription>
                Manage who has access to your organization
              </CardDescription>
            </div>
            <Button onClick={() => setShowInviteForm(true)}>
              <UserPlus className="h-4 w-4 mr-2" />
              Invite Member
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {/* Invite Form */}
          {showInviteForm && (
            <div className="mb-6 p-4 bg-slate-50 rounded-lg border border-slate-200">
              <form
                onSubmit={inviteForm.handleSubmit(onInviteSubmit)}
                className="space-y-4"
              >
                <div className="flex items-center justify-between">
                  <h4 className="font-medium text-slate-900">Invite New Member</h4>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => setShowInviteForm(false)}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="inviteEmail">Email Address</Label>
                    <div className="relative">
                      <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
                      <Input
                        id="inviteEmail"
                        type="email"
                        className="pl-10"
                        {...inviteForm.register("email")}
                        placeholder="colleague@company.com"
                      />
                    </div>
                    {inviteForm.formState.errors.email && (
                      <p className="text-sm text-red-600">
                        {inviteForm.formState.errors.email.message}
                      </p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="inviteRole">Role</Label>
                    <select
                      id="inviteRole"
                      {...inviteForm.register("role")}
                      className="w-full h-10 px-3 rounded-md border border-slate-300 bg-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="member">Member</option>
                      <option value="viewer">Viewer</option>
                      <option value="admin">Admin</option>
                    </select>
                  </div>
                </div>

                <div className="flex justify-end gap-3">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setShowInviteForm(false)}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={isInviting}>
                    {isInviting ? "Sending..." : "Send Invite"}
                  </Button>
                </div>
              </form>
            </div>
          )}

          {/* Members List */}
          <div className="space-y-3">
            {teamMembers.map((member) => (
              <div
                key={member.id}
                className="flex items-center justify-between p-4 bg-slate-50 rounded-lg"
              >
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-slate-200 flex items-center justify-center text-slate-600 font-medium">
                    {member.name.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <p className="font-medium text-slate-900">{member.name}</p>
                      {member.status === "pending" && (
                        <Badge variant="warning" className="text-xs">
                          Pending
                        </Badge>
                      )}
                      {member.id === currentUser?.id && (
                        <Badge variant="secondary" className="text-xs">
                          You
                        </Badge>
                      )}
                    </div>
                    <p className="text-sm text-slate-500">{member.email}</p>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <Badge className={roleConfig[member.role].color}>
                    {roleConfig[member.role].label}
                  </Badge>

                  {member.id !== currentUser?.id && (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() => handleChangeRole(member.id, "admin")}
                        >
                          <Shield className="h-4 w-4 mr-2" />
                          Make Admin
                          {member.role === "admin" && (
                            <Check className="h-4 w-4 ml-auto" />
                          )}
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => handleChangeRole(member.id, "member")}
                        >
                          Make Member
                          {member.role === "member" && (
                            <Check className="h-4 w-4 ml-auto" />
                          )}
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => handleChangeRole(member.id, "viewer")}
                        >
                          Make Viewer
                          {member.role === "viewer" && (
                            <Check className="h-4 w-4 ml-auto" />
                          )}
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          onClick={() => handleRemoveMember(member.id)}
                          className="text-red-600 focus:text-red-600"
                        >
                          <Trash2 className="h-4 w-4 mr-2" />
                          Remove
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                </div>
              </div>
            ))}
          </div>

          {/* Role Descriptions */}
          <div className="mt-6 pt-6 border-t border-slate-200">
            <h4 className="text-sm font-medium text-slate-700 mb-3">Role Permissions</h4>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              {(Object.keys(roleConfig) as Role[]).map((role) => (
                <div key={role} className="p-3 bg-slate-50 rounded-lg">
                  <Badge className={cn("mb-2", roleConfig[role].color)}>
                    {roleConfig[role].label}
                  </Badge>
                  <p className="text-xs text-slate-600">{roleConfig[role].description}</p>
                </div>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
