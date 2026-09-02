import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Loader2 } from 'lucide-react';
import { useState } from 'react';

interface Props {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    roles: any[];
    redirectUris: any[];
    onInvite: (data: { email: string; roleId?: string; redirectUri?: string }) => Promise<void>;
    isPending?: boolean;
}

export default function InviteUserDialog({
    open,
    onOpenChange,
    roles,
    redirectUris,
    onInvite,
    isPending = false,
}: Props) {
    const [email, setEmail] = useState('');
    const [selectedRole, setSelectedRole] = useState<string>('none');
    const [selectedRedirectUri, setSelectedRedirectUri] = useState<string>('none');
    const [error, setError] = useState<string>('');

    const handleInvite = async () => {
        setError('');

        // Validate email
        if (!email.trim()) {
            setError('Email is required');
            return;
        }

        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) {
            setError('Please enter a valid email address');
            return;
        }

        try {
            await onInvite({
                email: email.trim().toLowerCase(),
                roleId: selectedRole === 'none' ? undefined : selectedRole,
                redirectUri: selectedRedirectUri === 'none' ? undefined : selectedRedirectUri,
            });

            // Reset form on success
            setEmail('');
            setSelectedRole('none');
            setSelectedRedirectUri('none');
            onOpenChange(false);
        } catch (err: any) {
            setError(err.message || 'Failed to send invitation');
        }
    };

    const handleOpenChange = (newOpen: boolean) => {
        if (!newOpen) {
            // Reset form when closing
            setEmail('');
            setSelectedRole('none');
            setSelectedRedirectUri('none');
            setError('');
        }
        onOpenChange(newOpen);
    };

    return (
        <Dialog open={open} onOpenChange={handleOpenChange}>
            <DialogContent className="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle>Invite User</DialogTitle>
                    <DialogDescription>
                        Send an invitation email to a user to join this project.
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-4 py-2">
                    <div className="text-sm text-muted-foreground bg-muted/50 p-3 rounded border">
                        An invitation email will be sent to the user with a link to activate their account.
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="invite-email">Email</Label>
                        <Input
                            id="invite-email"
                            type="email"
                            placeholder="user@example.com"
                            value={email}
                            onChange={(e) => {
                                setEmail(e.target.value);
                                setError('');
                            }}
                            disabled={isPending}
                            className="w-full"
                        />
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="invite-role">Role (Optional)</Label>
                        <Select value={selectedRole} onValueChange={setSelectedRole} disabled={isPending}>
                            <SelectTrigger id="invite-role" className="w-full">
                                <SelectValue placeholder="Select a role..." />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="none">
                                    <span className="text-muted-foreground">No Role</span>
                                </SelectItem>
                                {roles.map((role) => (
                                    <SelectItem key={role.id} value={role.id}>
                                        {role.name}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="invite-redirect-uri">Login URL (Optional)</Label>
                        <Select value={selectedRedirectUri} onValueChange={setSelectedRedirectUri} disabled={isPending}>
                            <SelectTrigger id="invite-redirect-uri" className="w-full">
                                <SelectValue placeholder="Select login URL..." />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="none">
                                    <span className="text-muted-foreground">None</span>
                                </SelectItem>
                                {redirectUris.map((item, idx) => (
                                    <SelectItem key={`${item.redirect_uri || 'uri'}-${idx}`} value={item.redirect_uri}>
                                        {item.login_url && item.login_url.trim() ? item.login_url : item.redirect_uri}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                </div>

                <DialogFooter>
                    <Button
                        variant="outline"
                        onClick={() => handleOpenChange(false)}
                        disabled={isPending}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={handleInvite}
                        disabled={isPending || !email.trim()}
                    >
                        {isPending ? (
                            <>
                                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                Sending Invitation...
                            </>
                        ) : (
                            'Send Invitation'
                        )}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
