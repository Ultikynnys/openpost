<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { auth } from '$lib/stores/auth';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { instanceStore } from '$lib/stores/instance.svelte';
	import { recreateClient } from '$lib/api/client';
	import { IS_CAPACITOR } from '$lib/env';
	import { m } from '$lib/paraglide/messages';
	import {
		isNavigationItemActive,
		primaryNavigation,
		type PrimaryNavigationItem
	} from '$lib/app-navigation';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Avatar from '$lib/components/ui/avatar';
	import Logo from './Logo.svelte';
	import SidebarPlanner from './sidebar-planner.svelte';
	import LanguageSwitcher from './language-switcher.svelte';
	import CalendarIcon from 'lucide-svelte/icons/calendar-days';
	import ComposeIcon from 'lucide-svelte/icons/square-pen';
	import PostsIcon from 'lucide-svelte/icons/files';
	import MediaIcon from 'lucide-svelte/icons/images';
	import AccountsIcon from 'lucide-svelte/icons/users';
	import SettingsIcon from 'lucide-svelte/icons/settings';
	import UserIcon from 'lucide-svelte/icons/user-round';
	import PaletteIcon from 'lucide-svelte/icons/palette';
	import LogOutIcon from 'lucide-svelte/icons/log-out';
	import ChevronsUpDownIcon from 'lucide-svelte/icons/chevrons-up-down';
	import ServerIcon from 'lucide-svelte/icons/server';
	import CheckIcon from 'lucide-svelte/icons/check';
	import { setMode, userPrefersMode } from 'mode-watcher';
	import type { Workspace } from '$lib/api/client';
	type AppearanceMode = 'system' | 'light' | 'dark';

	let authState = $derived($auth);
	const sidebar = Sidebar.useSidebar();
	const currentPath = $derived(page.url.pathname);
	const currentWorkspaceName = $derived(workspaceCtx.currentWorkspace?.name ?? 'Select workspace');
	const currentWorkspaceAvatarURL = $derived(workspaceAvatarURL(workspaceCtx.currentWorkspace));
	const currentWorkspaceInitials = $derived(workspaceInitials(workspaceCtx.currentWorkspace));
	const userDisplayName = $derived(
		authState.user?.display_name || authState.user?.email?.split('@')[0] || m.common_untitled_user()
	);
	const userAvatarURL = $derived(authState.user?.avatar_url ?? '');
	const userInitials = $derived(initials(userDisplayName || authState.user?.email || 'User'));
	const navigationItems = $derived(
		primaryNavigation.map((item) => ({ ...item, icon: navigationIcon(item.id) }))
	);
	const newPostItem = $derived(navigationItems.find((item) => item.id === 'new')!);
	const workspaceNavigationItems = $derived(
		navigationItems.filter((item) => ['posts', 'media', 'accounts', 'settings'].includes(item.id))
	);
	const showDesktopPlanner = $derived(!sidebar.isMobile && sidebar.state === 'expanded');

	function navigationIcon(id: PrimaryNavigationItem['id']) {
		switch (id) {
			case 'new':
				return ComposeIcon;
			case 'calendar':
				return CalendarIcon;
			case 'posts':
				return PostsIcon;
			case 'media':
				return MediaIcon;
			case 'accounts':
				return AccountsIcon;
			default:
				return SettingsIcon;
		}
	}

	function initials(value: string) {
		const parts = value
			.replace(/@.*/, '')
			.split(/[\s._-]+/)
			.filter(Boolean);
		return ((parts[0]?.[0] ?? 'O') + (parts[1]?.[0] ?? '')).toUpperCase();
	}

	function workspaceAvatarURL(workspace: Workspace | null | undefined) {
		return (
			(workspace as (Workspace & { avatar_url?: string }) | null | undefined)?.avatar_url ?? ''
		).trim();
	}

	function workspaceInitials(workspace: Workspace | null | undefined) {
		return initials(workspace?.name || 'Workspace');
	}

	async function switchWorkspace(workspace: Workspace) {
		if (workspace.id === workspaceCtx.currentWorkspace?.id) return;
		await workspaceCtx.setWorkspace(workspace);
		sidebar.setOpenMobile(false);
	}

	function navigate(href: string) {
		sidebar.setOpenMobile(false);
		goto(resolve(href as '/'));
	}

	async function handleLogout() {
		await auth.logout();
		await goto(resolve('/login' as '/'));
	}

	async function handleSwitchServer() {
		await auth.logout();
		instanceStore().clearInstanceUrl();
		recreateClient();
		await goto(resolve('/connect' as '/'));
	}

	function chooseAppearance(nextMode: AppearanceMode) {
		setMode(nextMode);
	}
</script>

<Sidebar.Root collapsible="icon">
	<Sidebar.Header class="gap-2 border-b border-sidebar-border p-2" data-testid="app-sidebar">
		<a
			href={resolve('/')}
			class="flex h-10 items-center gap-2 rounded-md px-2 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0 focus-visible:ring-2 focus-visible:ring-sidebar-ring focus-visible:outline-none"
			aria-label="OpenPost home"
		>
			<Logo width={26} height={26} showText={sidebar.state !== 'collapsed'} />
		</a>

		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Sidebar.MenuButton
						{...props}
						size="lg"
						class="border border-sidebar-border bg-sidebar-accent/35 data-[state=open]:bg-sidebar-accent"
						tooltipContent="Switch workspace"
					>
						<Avatar.Root class="size-8 rounded-md">
							{#if currentWorkspaceAvatarURL}
								<Avatar.Image src={currentWorkspaceAvatarURL} alt={currentWorkspaceName} />
							{/if}
							<Avatar.Fallback class="rounded-md bg-primary/12 text-xs font-semibold text-primary">
								{currentWorkspaceInitials}
							</Avatar.Fallback>
						</Avatar.Root>
						<div class="grid min-w-0 flex-1 text-start leading-tight">
							<span class="truncate text-sm font-medium">{currentWorkspaceName}</span>
							<span class="truncate text-xs text-sidebar-foreground/62">Workspace</span>
						</div>
						<ChevronsUpDownIcon class="ms-auto size-4" />
					</Sidebar.MenuButton>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content
				class="w-64"
				side={sidebar.isMobile ? 'bottom' : 'right'}
				align="start"
				sideOffset={6}
			>
				<DropdownMenu.Label>Switch workspace</DropdownMenu.Label>
				{#each workspaceCtx.workspaces as workspace (workspace.id)}
					<DropdownMenu.Item onclick={() => switchWorkspace(workspace)} class="gap-3 py-2">
						<Avatar.Root class="size-8 rounded-md">
							{@const avatarURL = workspaceAvatarURL(workspace)}
							{#if avatarURL}<Avatar.Image src={avatarURL} alt={workspace.name} />{/if}
							<Avatar.Fallback class="rounded-md bg-muted text-xs">
								{workspaceInitials(workspace)}
							</Avatar.Fallback>
						</Avatar.Root>
						<span class="min-w-0 flex-1 truncate">{workspace.name}</span>
						{#if workspace.id === workspaceCtx.currentWorkspace?.id}
							<CheckIcon class="size-4 text-primary" />
						{/if}
					</DropdownMenu.Item>
				{/each}
				{#if workspaceCtx.workspaces.length === 0}
					<DropdownMenu.Item disabled>No workspaces</DropdownMenu.Item>
				{/if}
				<DropdownMenu.Separator />
				<DropdownMenu.Item onclick={() => navigate('/settings?tab=general')}>
					<SettingsIcon class="mr-2 size-4 text-muted-foreground" />
					Workspace settings
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.Header>

	<Sidebar.Content class={showDesktopPlanner ? 'py-2' : 'px-2 py-3'}>
		{#if showDesktopPlanner}
			<Sidebar.Group class="px-2 pt-0 pb-2">
				<Sidebar.GroupContent>
					<Sidebar.Menu>
						<Sidebar.MenuItem>
							<Sidebar.MenuButton
								isActive={isNavigationItemActive(newPostItem, currentPath)}
								class="h-10 bg-primary text-sm font-medium text-primary-foreground shadow-xs hover:bg-primary/90 hover:text-primary-foreground data-active:bg-primary data-active:text-primary-foreground"
								tooltipContent={newPostItem.label}
								onclick={() => navigate(newPostItem.href)}
							>
								<newPostItem.icon class="size-4" />
								<span>{newPostItem.label}</span>
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					</Sidebar.Menu>
				</Sidebar.GroupContent>
			</Sidebar.Group>

			<SidebarPlanner onNavigate={navigate} />
		{:else}
			<Sidebar.Group class="p-0">
				<Sidebar.GroupLabel
					class="px-2 text-[11px] tracking-[0.12em] text-sidebar-foreground/48 uppercase"
				>
					Publish
				</Sidebar.GroupLabel>
				<Sidebar.GroupContent>
					<Sidebar.Menu class="gap-1">
						{#each navigationItems as item (item.id)}
							<Sidebar.MenuItem>
								<Sidebar.MenuButton
									isActive={isNavigationItemActive(item, currentPath)}
									class={item.id === 'new'
										? 'h-10 bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground data-active:bg-primary data-active:text-primary-foreground'
										: 'h-10 text-sm'}
									tooltipContent={item.label}
									onclick={() => navigate(item.href)}
								>
									<item.icon class="size-4" />
									<span>{item.label}</span>
								</Sidebar.MenuButton>
							</Sidebar.MenuItem>
						{/each}
					</Sidebar.Menu>
				</Sidebar.GroupContent>
			</Sidebar.Group>
		{/if}
	</Sidebar.Content>

	<Sidebar.Footer class="border-t border-sidebar-border p-2">
		{#if showDesktopPlanner}
			<div class="pb-2">
				<p
					class="flex h-7 items-center px-2 text-[11px] tracking-[0.1em] text-sidebar-foreground/52 uppercase"
				>
					Workspace
				</p>
				<Sidebar.Menu class="grid grid-cols-2 gap-1">
					{#each workspaceNavigationItems as item (item.id)}
						<Sidebar.MenuItem>
							<Sidebar.MenuButton
								isActive={isNavigationItemActive(item, currentPath)}
								class="h-9 gap-1.5 px-2 text-xs"
								tooltipContent={item.label}
								onclick={() => navigate(item.href)}
							>
								<item.icon class="size-3.5" />
								<span>{item.id === 'accounts' ? 'Accounts' : item.label}</span>
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					{/each}
				</Sidebar.Menu>
			</div>
		{/if}
		<Sidebar.Menu class={showDesktopPlanner ? 'border-t border-sidebar-border pt-2' : ''}>
			<Sidebar.MenuItem>
				<DropdownMenu.Root>
					<DropdownMenu.Trigger>
						{#snippet child({ props })}
							<Sidebar.MenuButton
								{...props}
								size="lg"
								class="data-[state=open]:bg-sidebar-accent"
								tooltipContent="Profile and appearance"
							>
								<Avatar.Root class="size-8 rounded-full">
									{#if userAvatarURL}<Avatar.Image src={userAvatarURL} alt={userDisplayName} />{/if}
									<Avatar.Fallback
										class="bg-sidebar-primary text-xs text-sidebar-primary-foreground"
									>
										{userInitials}
									</Avatar.Fallback>
								</Avatar.Root>
								<div class="grid min-w-0 flex-1 text-start leading-tight">
									<span class="truncate text-sm font-medium">{userDisplayName}</span>
									<span class="truncate text-xs text-sidebar-foreground/62"
										>{authState.user?.email}</span
									>
								</div>
								<ChevronsUpDownIcon class="ms-auto size-4" />
							</Sidebar.MenuButton>
						{/snippet}
					</DropdownMenu.Trigger>
					<DropdownMenu.Content
						class="w-60"
						side={sidebar.isMobile ? 'bottom' : 'right'}
						align="end"
						sideOffset={6}
					>
						<DropdownMenu.Item onclick={() => navigate('/settings?tab=profile')}>
							<UserIcon class="mr-2 size-4 text-muted-foreground" />
							Profile & security
						</DropdownMenu.Item>
						<DropdownMenu.Sub>
							<DropdownMenu.SubTrigger>
								<PaletteIcon class="mr-2 size-4 text-muted-foreground" />
								Appearance
								<span class="ml-auto text-muted-foreground capitalize"
									>{userPrefersMode.current}</span
								>
							</DropdownMenu.SubTrigger>
							<DropdownMenu.SubContent class="w-40">
								{#each ['system', 'light', 'dark'] as appearance (appearance)}
									<DropdownMenu.Item onclick={() => chooseAppearance(appearance as AppearanceMode)}>
										<span class="capitalize">{appearance}</span>
										{#if userPrefersMode.current === appearance}
											<CheckIcon class="ml-auto size-4 text-primary" />
										{/if}
									</DropdownMenu.Item>
								{/each}
							</DropdownMenu.SubContent>
						</DropdownMenu.Sub>
						<LanguageSwitcher variant="menu" />
						{#if IS_CAPACITOR}
							<DropdownMenu.Separator />
							<DropdownMenu.Item onclick={handleSwitchServer}>
								<ServerIcon class="mr-2 size-4 text-muted-foreground" />
								{m.sidebar_change_server()}
							</DropdownMenu.Item>
						{/if}
						<DropdownMenu.Separator />
						<DropdownMenu.Item onclick={handleLogout}>
							<LogOutIcon class="mr-2 size-4 text-muted-foreground" />
							{m.sidebar_log_out()}
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Sidebar.MenuItem>
		</Sidebar.Menu>
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
