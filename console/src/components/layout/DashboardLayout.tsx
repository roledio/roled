import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "./AppSidebar";
import type { HttpClient } from "@/services/core/httpClient";
import type { TokenService } from "@/services/core/tokenService";
import { useCurrentTokenAndMemberInfo, useRevokeToken } from "@/hooks/use-current-token-info";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import { LogOut, User, Book } from "lucide-react";
import { useNavigate } from 'react-router-dom';
import { telemetry } from "@/lib/telemetry";
import React, { useEffect, useRef } from "react";

interface DashboardLayoutProps {
  children: React.ReactNode;
  httpClient: HttpClient;
  tokenService: TokenService;
}

export function DashboardLayout({ children, httpClient, tokenService }: DashboardLayoutProps) {
  const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;
  const DOCS_BASE_URL = (import.meta.env.VITE_DOCS_BASE_URL as string) || "";
  const { data: info, isLoading } = useCurrentTokenAndMemberInfo({ httpClient, authBaseUrl: AUTH_BASE_URL });
  const tokenInfo = info?.tokenInfo;
  const navigate = useNavigate();
  const identifiedUserId = useRef<string | null>(null);

  useEffect(() => {
    const user = tokenInfo?.user;
    if (!user?.id || identifiedUserId.current === user.id) return;

    telemetry.identifyUser(user.id, {
      email: user.email,
      name: user.display_name,
      role: tokenInfo.role.name,
    });
    identifiedUserId.current = user.id;
  }, [tokenInfo]);

  const revokeMutation = useRevokeToken({
    httpClient,
    authBaseUrl: AUTH_BASE_URL,
    onSuccess: () => {
      tokenService.clear();
      window.location.href = "/signin";
    },
    onError: () => {
      // Clear and redirect anyway so user doesn't get stuck
      tokenService.clear();
      window.location.href = "/signin";
    },
  });

  const handleSignOut = () => {
    telemetry.resetUser();
    const clientId = tokenInfo?.client?.id;
    const refreshToken = tokenService.getRefreshToken();
    if (clientId && refreshToken) {
      revokeMutation.mutate({
        client_id: clientId,
        refresh_token: refreshToken,
      });
    } else {
      tokenService.clear();
      window.location.href = "/signin";
    }
  };

  const getAvatarLetter = (name?: string) => {
    if (!name) return "?";
    const firstName = name.trim().split(/\s+/)[0];
    return firstName ? firstName.charAt(0).toUpperCase() : "?";
  };

  const displayName = tokenInfo?.user?.display_name || "User";
  const email = tokenInfo?.user?.email || "";
  const avatarLetter = getAvatarLetter(displayName);

  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        <AppSidebar />
        <div className="flex-1 flex flex-col min-w-0">
          <header className="h-12 flex items-center border-b bg-card px-4 shrink-0 justify-between">
            <SidebarTrigger />
            <div className="ml-auto flex items-center gap-4">
              {DOCS_BASE_URL && (
                <a
                  href={DOCS_BASE_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-2 px-2 py-1 rounded outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-all duration-200 hover:scale-105 active:scale-95"
                  aria-label="Docs"
                >
                  <Book className="h-4 w-4" />
                  <span className="hidden sm:inline text-sm">Docs</span>
                </a>
              )}
              {isLoading ? (
                <div className="h-8 w-8 rounded-full bg-muted animate-pulse" />
              ) : (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      className="flex items-center gap-2 outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded transition-all duration-200 hover:scale-105 active:scale-95"
                      aria-label="User menu"
                      disabled={revokeMutation.status === "pending"}
                    >
                      {tokenInfo?.user?.avatar_url ? (
                        <img
                          src={tokenInfo.user.avatar_url}
                          alt={displayName}
                          className="h-8 w-8 rounded object-cover border border-border"
                        />
                      ) : (
                        <div className="h-8 w-8 rounded bg-muted flex items-center justify-center text-sm font-bold text-muted-foreground">{avatarLetter}</div>
                      )}
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent className="w-56" align="end" forceMount>
                    <DropdownMenuLabel className="font-normal">
                      <div className="flex flex-col space-y-1">
                        <p className="text-sm font-semibold leading-none text-foreground">{displayName}</p>
                        <p className="text-xs leading-none text-muted-foreground">{email}</p>
                      </div>
                    </DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      onClick={() => navigate('/profile')}
                      className="cursor-pointer flex items-center gap-2"
                    >
                      <User className="h-4 w-4" />
                      <span>Profile</span>
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      onClick={handleSignOut}
                      className="text-destructive focus:text-destructive focus:bg-destructive/10 cursor-pointer flex items-center gap-2"
                      disabled={revokeMutation.status === "pending"}
                    >
                      <LogOut className="h-4 w-4" />
                      <span>{revokeMutation.status === "pending" ? "Signing Out..." : "Sign Out"}</span>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
            </div>
          </header>
          <main className="flex-1 p-4 md:p-6 overflow-auto">
            {children}
          </main>
        </div>
      </div>
    </SidebarProvider>
  );
}
