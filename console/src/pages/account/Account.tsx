import ApiPagination from '@/components/ApiPagination';
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { StatusBadge } from "@/components/StatusBadge";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from "@/components/ui/textarea";
import { useCurrentAccount, useDeleteAccount } from "@/hooks/accounts";
import { useDeleteMember, useInviteMember, useMembers, useUpdateMember } from '@/hooks/members';
import { useToast } from '@/hooks/use-toast';
import { useCurrentTokenAndMemberInfo } from "@/hooks/use-current-token-info";
import { formatDate } from '@/lib/date';
import { updateAccount } from '@/services/accounts';
import type { HttpClient } from "@/services/core/httpClient";
import type { TokenService } from "@/services/core/tokenService";
import { useQueryClient } from '@tanstack/react-query';
import { telemetry } from '@/lib/telemetry';
import { AlertCircle, CheckCircle2, ChevronDown, ChevronUp, Eye, EyeOff, Loader2, MoreHorizontal, MoreVertical, Search, Shield, Trash2, XCircle } from "lucide-react";
import { useEffect, useState } from "react";
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

type AccountDraft = {
  name: string;
  description: string;
};

interface AccountProps {
  httpClient: HttpClient;
  tokenService: TokenService;
}

export default function Account({ httpClient, tokenService }: AccountProps) {
  const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;

  // Fetch account from API
  const { account, isLoading, error } = useCurrentAccount({
    baseUrl: AUTH_BASE_URL,
    httpClient,
  });

  // Fetch current token and member info to check logged-in user email and admin status
  const { data: currentInfo } = useCurrentTokenAndMemberInfo({
    httpClient,
    authBaseUrl: AUTH_BASE_URL,
  });

  const currentUserEmail = currentInfo?.tokenInfo?.user?.email;
  const isCurrentUserAdmin = currentInfo?.memberInfo?.is_admin === true;

  // Edit state
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState<AccountDraft>({ name: "", description: "" });

  const queryClient = useQueryClient();
  const { toast } = useToast();

  // Sync draft when account data arrives
  useEffect(() => {
    if (account) {
      setDraft({ name: account.name, description: account.description });
    }
  }, [account]);

  // Delete account
  const [showDelete, setShowDelete] = useState(false);
  const [deletePassword, setDeletePassword] = useState("");
  const [showDeletePassword, setShowDeletePassword] = useState(false);
  const [deletingAccount, setDeletingAccount] = useState(false);

  const deleteAccountMutation = useDeleteAccount({ httpClient, baseUrl: AUTH_BASE_URL, accountId: account?.id });

  const handleDeleteAccount = async () => {
    if (!deletePassword) return;
    setDeletingAccount(true);
    try {
      await deleteAccountMutation.mutateAsync({ password: deletePassword });
      // Clear all browser data and redirect to signin
      telemetry.resetUser();
      tokenService.clear();
      window.location.href = '/signin';
    } catch (err: any) {
      const apiMessage = err?.response?.data?.error?.message || err?.message || 'Failed to delete account';
      toast({ title: 'Delete failed', description: apiMessage, variant: 'destructive' });
    } finally {
      setDeletingAccount(false);
    }
  };

  // Members (UI shape)
  type MemberRow = {
    id: string;
    email: string;
    displayName: string;
    isActive: boolean;
    emailVerified: boolean;
    role: 'admin' | 'member';
    updatedAt?: string | null;
    createdAt?: string | null;
    avatarUrl?: string | null;
  };
  const [removeMember, setRemoveMember] = useState<MemberRow | null>(null);

  // filters and pagination
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [isVerified, setIsVerified] = useState<string | null>(null);
  const [isActive, setIsActive] = useState<string | null>(null);
  const [isAdmin, setIsAdmin] = useState<string | null>(null);
  const [pageNum, setPageNum] = useState(1);
  const [pageSize, setPageSize] = useState<number>(5);
  // default sort: Last Updated DESC
  const [sortBy, setSortBy] = useState<string | null>('created_at');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

  useEffect(() => { const t = setTimeout(() => setSearch(searchInput), 300); return () => clearTimeout(t); }, [searchInput]);

  const [saving, setSaving] = useState(false);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteLoading, setInviteLoading] = useState(false);

  const doSaveAccount = async () => {
    if (!account) return;
    setSaving(true);
    try {
      const updated = await updateAccount(httpClient, AUTH_BASE_URL, account.id, {
        name: draft.name,
        description: draft.description,
      });

      // update react-query cache so UI reflects updated account
      queryClient.setQueryData(['account', 'current'], updated);
      toast({ title: 'Account updated', description: 'Account updated successfully' });
      setEditing(false);
    } catch (err: any) {
      const apiMessage =
        err?.response?.data?.error?.message || err?.response?.data?.message || err?.message || 'Failed to update account';
      toast({ title: 'Update failed', description: apiMessage, variant: 'destructive' });
    } finally {
      setSaving(false);
    }
  };

  const cancelEdit = () => {
    if (account) {
      setDraft({ name: account.name, description: account.description });
    }
    setEditing(false);
  };

  const startEdit = () => {
    if (account) {
      setDraft({ name: account.name, description: account.description });
    }
    setEditing(true);
  };

  const membersQuery = useMembers({
    httpClient,
    baseUrl: AUTH_BASE_URL,
    accountId: account?.id ?? null,
    pageNum,
    pageSize,
    search,
    isVerified,
    isActive,
    isAdmin,
    sortBy,
    sortDir,
  });

  const inviteMutation = useInviteMember({ httpClient, baseUrl: AUTH_BASE_URL, accountId: account?.id });
  const deleteMutation = useDeleteMember({ httpClient, baseUrl: AUTH_BASE_URL, accountId: account?.id });
  const updateMutation = useUpdateMember({ httpClient, baseUrl: AUTH_BASE_URL, accountId: account?.id });

  const confirmRemoveMember = async () => {
    if (!removeMember) return;
    try {
      await deleteMutation.mutateAsync(removeMember.id);
      toast({ title: 'Member removed', description: `${removeMember.displayName} removed successfully` });
    } catch (err: any) {
      const apiMessage = err?.response?.data?.error?.message || err?.message || 'Failed to remove member';
      toast({ title: 'Remove failed', description: apiMessage, variant: 'destructive' });
    } finally {
      setRemoveMember(null);
    }
  };

  const handleInviteConfirm = async () => {
    try {
      await inviteMutation.mutateAsync(inviteEmail);
      toast({ title: 'Invitation sent', description: `Invitation sent to ${inviteEmail}` });
    } catch (err: any) {
      const apiMessage = err?.response?.data?.error?.message || err?.message || 'Failed to invite member';
      toast({ title: 'Invite failed', description: apiMessage, variant: 'destructive' });
    } finally {
      setInviteOpen(false);
      setInviteEmail('');
    }
  };

  // members data from react-query
  const membersData = membersQuery.data?.data ?? [];
  const members = membersData.map((m: any) => ({
    id: m.id,
    email: m.email,
    displayName: m.display_name ?? m.displayName ?? m.email,
    isActive: (m as any).is_active ?? (m as any).isActive ?? false,
    emailVerified: (m as any).is_verified ?? (m as any).emailVerified ?? false,
    role: (m as any).is_admin ? 'admin' : 'member',
    updatedAt: (m as any).updated_at ?? (m as any).updatedAt ?? null,
    createdAt: (m as any).created_at ?? (m as any).createdAt ?? null,
    avatarUrl: m.avatar_url ?? m.avatarUrl ?? null,
  } as MemberRow));

  const membersPagination = membersQuery.data?.pagination ?? null;
  const membersLoading = membersQuery.isLoading || membersQuery.isFetching;

  // Loading state
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20" data-testid="account-loading">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        <span className="ml-2 text-sm text-muted-foreground">Loading account…</span>
      </div>
    );
  }

  // Error state
  if (error || !account) {
    return (
      <div className="flex flex-col items-center justify-center py-20 space-y-2" data-testid="account-error">
        <AlertCircle className="h-8 w-8 text-destructive" />
        <p className="text-sm text-destructive font-medium">Failed to load account</p>
        <p className="text-xs text-muted-foreground">{error?.message ?? "Account data is unavailable"}</p>
      </div>
    );
  }

  return (
    <div className="space-y-4 max-w-4xl">
      {/* Account Info */}
      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold">Account</h1>
            <p className="text-sm text-muted-foreground mt-1">Manage your account settings</p>
          </div>
          {!editing && (
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={startEdit}>
                Edit
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="text-destructive hover:text-destructive border-destructive/30 hover:bg-destructive/10"
                onClick={() => setShowDelete(true)}
              >
                Delete Account
              </Button>
            </div>
          )}
        </div>

        <div className="border rounded bg-card p-5 space-y-4">
          {editing ? (
            <>
              <div className="space-y-1.5">
                <Label htmlFor="account-name">Name</Label>
                <Input
                  id="account-name"
                  value={draft.name}
                  onChange={(e) => setDraft({ ...draft, name: e.target.value })}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="account-desc">Description</Label>
                <Textarea
                  id="account-desc"
                  value={draft.description}
                  onChange={(e) => setDraft({ ...draft, description: e.target.value })}
                  rows={3}
                />
              </div>
              <div className="flex gap-2 justify-end">
                <Button variant="outline" size="sm" onClick={cancelEdit}>
                  Cancel
                </Button>
                <Button size="sm" onClick={doSaveAccount} disabled={saving}>
                  {saving ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin mr-2" />
                      Saving...
                    </>
                  ) : (
                    <>
                      Save
                    </>
                  )}
                </Button>
              </div>
            </>
          ) : (
            <>
              <div>
                <p className="text-xs text-muted-foreground">Account Name</p>
                <p className="text-sm font-medium mt-0.5">{account.name}</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Description</p>
                <p className="text-sm mt-0.5">{account?.description || "-"}</p>
              </div>
              <div className="flex gap-8">
                <div>
                  <p className="text-xs text-muted-foreground">Status</p>
                  <div className="mt-0.5">
                    <StatusBadge active={account.is_active} />
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
      </section>

      {/* Members */}
      <section className="border rounded p-4 bg-card w-full space-y-4">
        <div>
          <h2 className="text-lg font-semibold">Members</h2>
          <div className="text-sm text-muted-foreground">Manage members of your account</div>
        </div>

        {/* Filters toolbar under title/subtitle */}
        <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center">
          <div className="relative flex-1 w-full sm:max-w-xs">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search members…"
              value={searchInput}
              onChange={(e) => { setSearchInput(e.target.value); setPageNum(1); }}
              className="pl-8"
              aria-label="Search members"
            />
          </div>

          <Select value={isVerified ?? 'all'} onValueChange={(v) => { setIsVerified(v === 'all' ? null : v); setPageNum(1); }}>
            <SelectTrigger className="w-[160px]" aria-label="Filter by verified">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="true">Verified</SelectItem>
              <SelectItem value="false">Not Verified</SelectItem>
            </SelectContent>
          </Select>

          <Select value={isActive ?? 'all'} onValueChange={(v) => { setIsActive(v === 'all' ? null : v); setPageNum(1); }}>
            <SelectTrigger className="w-[160px]" aria-label="Filter by active">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="true">Active</SelectItem>
              <SelectItem value="false">Inactive</SelectItem>
            </SelectContent>
          </Select>

          <Select value={isAdmin ?? 'all'} onValueChange={(v) => { setIsAdmin(v === 'all' ? null : v); setPageNum(1); }}>
            <SelectTrigger className="w-[160px]" aria-label="Filter by role">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="true">Admin</SelectItem>
              <SelectItem value="false">Member</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="flex items-center justify-between mb-2">
          <div className="text-sm text-muted-foreground">
            {membersPagination ? (
              membersPagination.page_size === membersPagination.total_data
                ? `Showing ${membersPagination.page_size} members`
                : `Showing ${membersPagination.page_size} of ${membersPagination.total_data} members`
            ) : null}
          </div>
          <div>
            <Button size="sm" onClick={() => setInviteOpen(true)}>
              Invite Member
            </Button>
          </div>
        </div>

        <div className="border rounded bg-card shadow-sm overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/40">
                  <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                    <button onClick={() => {
                      const key = 'name';
                      if (sortBy !== key) { setSortBy(key); setSortDir('asc'); } else { setSortDir(sortDir === 'asc' ? 'desc' : 'asc'); }
                      setPageNum(1);
                    }} className="flex items-center gap-2">
                      <span>Name</span>
                      {sortBy === 'name' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                    </button>
                  </th>
                  <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                    <button onClick={() => {
                      const key = 'is_active';
                      if (sortBy !== key) { setSortBy(key); setSortDir('asc'); } else { setSortDir(sortDir === 'asc' ? 'desc' : 'asc'); }
                      setPageNum(1);
                    }} className="flex items-center gap-2">
                      <span>Status</span>
                      {sortBy === 'is_active' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                    </button>
                  </th>
                  <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                    <button onClick={() => {
                      const key = 'is_admin';
                      if (sortBy !== key) { setSortBy(key); setSortDir('asc'); } else { setSortDir(sortDir === 'asc' ? 'desc' : 'asc'); }
                      setPageNum(1);
                    }} className="flex items-center gap-2">
                      <span>Role</span>
                      {sortBy === 'is_admin' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                    </button>
                  </th>
                  <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                    <button onClick={() => {
                      const key = 'created_at';
                      if (sortBy !== key) { setSortBy(key); setSortDir('asc'); } else { setSortDir(sortDir === 'asc' ? 'desc' : 'asc'); }
                      setPageNum(1);
                    }} className="flex items-center gap-2">
                      <span>Created</span>
                      {sortBy === 'created_at' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                    </button>
                  </th>
                  <th className="w-10 px-4 py-2.5"> </th>
                </tr>
              </thead>
              <tbody>
                {membersLoading ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        <span>Loading members…</span>
                      </div>
                    </td>
                  </tr>
                ) : members.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-3 text-sm text-muted-foreground">No members found</td>
                  </tr>
                ) : (
                  members.map((member) => (
                    <tr key={member.id} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-3">
                          {member.avatarUrl ? (
                            <img src={member.avatarUrl} alt={member.displayName} className="h-8 w-8 rounded object-cover border border-border" />
                          ) : (
                            <div className="h-8 w-8 rounded bg-muted flex items-center justify-center text-xs font-bold text-muted-foreground">
                              {(member.displayName ?? '')?.charAt(0).toUpperCase()}
                            </div>
                          )}
                          <div>
                            <div className="font-medium">{member.displayName}</div>
                            <div className="text-xs text-muted-foreground mt-1">
                              {member.email}
                              {member.emailVerified ? (
                                <span className="inline-flex items-center gap-1 text-success text-xs ml-2">
                                  <CheckCircle2 className="h-3.5 w-3.5" /> Verified
                                </span>
                              ) : (
                                <span className="inline-flex items-center gap-1 text-muted-foreground text-xs ml-2">
                                  <XCircle className="h-3.5 w-3.5" /> Not Verified
                                </span>
                              )}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <StatusBadge active={member.isActive} />
                      </td>
                      <td className="px-4 py-3">
                        <Badge variant={member.role === "admin" ? "default" : "secondary"} className="text-xs capitalize">
                          {member.role === 'admin' ? 'Admin' : 'Member'}
                        </Badge>
                      </td>
                      <td className="px-4 py-3 text-xs text-muted-foreground">
                        {member.createdAt ? formatDate(member.createdAt) : ''}
                      </td>
                      <td className="px-4 py-3">
                        {isCurrentUserAdmin && currentUserEmail && member.email !== currentUserEmail ? (
                          <DropdownMenu>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <DropdownMenuTrigger asChild>
                                  <Button
                                    size="sm"
                                    variant="ghost"
                                    aria-label={`Actions for ${member.displayName}`}
                                    className="h-8 w-8 p-0"
                                  >
                                    <MoreVertical className="h-4 w-4" />
                                  </Button>
                                </DropdownMenuTrigger>
                              </TooltipTrigger>
                              <TooltipContent>Actions</TooltipContent>
                            </Tooltip>
                            <DropdownMenuContent align="end">
                              {member.role === 'admin' ? (
                                <DropdownMenuItem
                                  onSelect={async () => {
                                    try {
                                      await updateMutation.mutateAsync({ memberId: member.id, body: { is_admin: false } });
                                      toast({ title: 'Role updated', description: `${member.displayName} is no longer an admin.` });
                                    } catch (err: any) {
                                      const apiMessage = err?.response?.data?.error?.message || err?.message || 'Failed to update member';
                                      toast({ title: 'Update failed', description: apiMessage, variant: 'destructive' });
                                    }
                                  }}
                                >
                                  <Shield className="h-4 w-4 mr-2 text-muted-foreground" />
                                  Remove as admin
                                </DropdownMenuItem>
                              ) : (
                                <DropdownMenuItem
                                  onSelect={async () => {
                                    try {
                                      await updateMutation.mutateAsync({ memberId: member.id, body: { is_admin: true } });
                                      toast({ title: 'Role updated', description: `${member.displayName} is now an admin.` });
                                    } catch (err: any) {
                                      const apiMessage = err?.response?.data?.error?.message || err?.message || 'Failed to update member';
                                      toast({ title: 'Update failed', description: apiMessage, variant: 'destructive' });
                                    }
                                  }}
                                >
                                  <Shield className="h-4 w-4 mr-2 text-muted-foreground" />
                                  Set as admin
                                </DropdownMenuItem>
                              )}
                              <DropdownMenuItem
                                className="text-destructive focus:text-destructive"
                                onSelect={() => setRemoveMember(member)}
                              >
                                <Trash2 className="h-4 w-4 mr-2" />
                                Remove
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        ) : (
                          <></>
                        )}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Select value={String(pageSize)} onValueChange={(v) => { setPageSize(Number(v)); setPageNum(1); }}>
              <SelectTrigger className="w-[96px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {[5, 25, 50].map((s) => <SelectItem key={s} value={String(s)}>{String(s)}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          <div>
            <ApiPagination
              pagination={membersPagination}
              pageSize={pageSize}
              onPageChange={setPageNum}
            />
          </div>
        </div>
      </section>

      {/* Delete Account Dialog */}
      <ConfirmDialog
        open={showDelete}
        onOpenChange={(open) => {
          setShowDelete(open);
          if (!open) {
            setDeletePassword("");
            setShowDeletePassword(false);
          }
        }}
        title="Delete Account"
        description="This will permanently delete your account and all associated data. Enter your password to confirm."
        confirmLabel={deletingAccount ? (<><Loader2 className="h-4 w-4 animate-spin mr-2" />Deleting...</>) : 'Delete Account'}
        destructive
        closeOnConfirm={false}
        disabled={deletePassword.length === 0 || deletingAccount}
        onConfirm={handleDeleteAccount}
      >
        <div className="py-2">
          <Label htmlFor="delete-password" className="text-sm">Password</Label>
          <div className="relative mt-1.5">
            <Input
              id="delete-password"
              type={showDeletePassword ? 'text' : 'password'}
              placeholder="Enter your password"
              value={deletePassword}
              onChange={(e) => setDeletePassword(e.target.value)}
              className="pr-10"
            />
            <button
              aria-label={showDeletePassword ? 'Hide password' : 'Show password'}
              onClick={() => setShowDeletePassword(s => !s)}
              className="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded text-sm text-muted-foreground hover:bg-muted/5"
              type="button"
            >
              {showDeletePassword ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
            </button>
          </div>
        </div>
      </ConfirmDialog>

      {/* Remove Member Dialog */}
      <ConfirmDialog
        open={!!removeMember}
        onOpenChange={(open) => !open && setRemoveMember(null)}
        title="Remove Member"
        description={<span>Are you sure you want to remove <span className="font-medium">{removeMember?.displayName}</span>? They will lose access immediately.</span>}
        confirmLabel={deleteMutation.status === 'pending' ? (<><Loader2 className="h-4 w-4 animate-spin mr-2" />Removing...</>) : 'Remove'}
        destructive
        closeOnConfirm={false}
        disabled={deleteMutation.status === 'pending'}
        onConfirm={confirmRemoveMember}
      />

      {/* Invite Member Dialog */}
      <ConfirmDialog
        open={inviteOpen}
        onOpenChange={(open) => {
          setInviteOpen(open);
          if (!open) setInviteEmail('');
        }}
        title="Invite Member"
        description="Enter the email address of the user you want to invite. An invitation email will be sent."
        confirmLabel={inviteMutation.status === 'pending' ? (<><Loader2 className="h-4 w-4 animate-spin mr-2" />Inviting...</>) : 'Invite'}
        closeOnConfirm={false}
        disabled={!inviteEmail || !inviteEmail.includes('@') || inviteMutation.status === 'pending'}
        onConfirm={handleInviteConfirm}
      >
        <div className="py-2">
          <Label htmlFor="invite-email" className="text-sm">Email</Label>
          <Input
            id="invite-email"
            type="email"
            placeholder="email@example.com"
            value={inviteEmail}
            onChange={(e) => setInviteEmail(e.target.value)}
            className="mt-1.5"
          />
        </div>
      </ConfirmDialog>
    </div>
  );
}
