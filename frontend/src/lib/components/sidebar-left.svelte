<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import { auth } from '$lib/stores/auth';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { ui } from '$lib/stores/ui.svelte';
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
	import AccountPreferencesMenu from './account-preferences-menu.svelte';
	import CalendarIcon from 'lucide-svelte/icons/calendar-days';
	import ComposeIcon from 'lucide-svelte/icons/square-pen';
	import PostsIcon from 'lucide-svelte/icons/files';
	import MediaIcon from 'lucide-svelte/icons/images';
	import AccountsIcon from 'lucide-svelte/icons/users';
	import SettingsIcon from 'lucide-svelte/icons/settings';
	import ChevronsUpDownIcon from 'lucide-svelte/icons/chevrons-up-down';
	import CheckIcon from 'lucide-svelte/icons/check';
	import type { Workspace } from '$lib/api/client';

	let authState = $derived($auth);
	const sidebar = Sidebar.useSidebar();
	const currentPath = $derived(page.url.pathname);
	const currentWorkspaceName = $derived(
		workspaceCtx.currentWorkspace?.name ?? m.sidebar_select_workspace()
	);
	const currentWorkspaceAvatarURL = $derived(workspaceAvatarURL(workspaceCtx.currentWorkspace));
	const currentWorkspaceInitials = $derived(workspaceInitials(workspaceCtx.currentWorkspace));
	const userDisplayName = $derived(
		authState.user?.display_name || authState.user?.email?.split('@')[0] || m.common_untitled_user()
	);
	const userAvatarURL = $derived(authState.user?.avatar_url ?? '');
	const userInitials = $derived(initials(userDisplayName || authState.user?.email || 'User'));
	const navigationItems = $derived(
		primaryNavigation.map((item) => ({
			...item,
			label: navigationLabel(item.id),
			icon: navigationIcon(item.id)
		}))
	);
	const sidebarNavigationItems = $derived(navigationItems.filter((item) => item.id !== 'new'));
	const workspaceNavigationItems = $derived(
		navigationItems.filter((item) => ['posts', 'media', 'accounts', 'settings'].includes(item.id))
	);
	const showDesktopPlanner = $derived(!sidebar.isMobile && sidebar.state === 'expanded');
	const showHomeBrand = $derived(currentPath === '/' && !ui.activeComposerDraftId);

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

	function navigationLabel(id: PrimaryNavigationItem['id']) {
		switch (id) {
			case 'new':
				return m.sidebar_new_post();
			case 'calendar':
				return m.sidebar_calendar();
			case 'posts':
				return m.sidebar_activity();
			case 'media':
				return m.sidebar_media();
			case 'accounts':
				return m.sidebar_accounts();
			case 'settings':
				return m.sidebar_settings();
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
		if (href === '/') ui.startNewPost();
		goto(resolve(href as '/'));
	}
</script>

<Sidebar.Root collapsible="icon">
	<Sidebar.Header class="gap-2 border-b border-sidebar-border p-2" data-testid="app-sidebar">
		<div class="relative h-10 overflow-hidden rounded-md">
			{#key showHomeBrand}
				{#if showHomeBrand}
					<a
						href={resolve('/')}
						class="absolute inset-0 flex items-center gap-2 rounded-md px-2 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0 focus-visible:ring-2 focus-visible:ring-sidebar-ring focus-visible:outline-none"
						aria-label={m.sidebar_openpost_home()}
						data-testid="sidebar-home-brand"
						transition:fly={{ x: -12, duration: 180, easing: quintOut }}
					>
						<Logo width={26} height={26} showText={sidebar.state !== 'collapsed'} />
					</a>
				{:else}
					<a
						href={resolve('/')}
						class="absolute inset-0 flex items-center gap-2 rounded-md bg-primary px-3 text-sm font-semibold text-primary-foreground shadow-xs group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0 hover:bg-primary/90 focus-visible:ring-2 focus-visible:ring-sidebar-ring focus-visible:outline-none"
						aria-label={m.sidebar_new_post()}
						data-testid="sidebar-new-post"
						onclick={() => ui.startNewPost()}
						transition:fly={{ x: 12, duration: 180, easing: quintOut }}
					>
						<ComposeIcon class="size-4" />
						{#if sidebar.state !== 'collapsed'}<span>{m.sidebar_new_post()}</span>{/if}
					</a>
				{/if}
			{/key}
		</div>

		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Sidebar.MenuButton
						{...props}
						size="lg"
						class="border border-sidebar-border bg-sidebar-accent/35 data-[state=open]:bg-sidebar-accent"
						tooltipContent={m.sidebar_switch_workspace()}
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
							<span class="truncate text-xs text-sidebar-foreground/62"
								>{m.sidebar_workspace()}</span
							>
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
				<DropdownMenu.Label>{m.sidebar_switch_workspace()}</DropdownMenu.Label>
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
					<DropdownMenu.Item disabled>{m.sidebar_no_workspaces()}</DropdownMenu.Item>
				{/if}
				<DropdownMenu.Separator />
				<DropdownMenu.Item onclick={() => navigate('/settings?tab=general')}>
					<SettingsIcon class="mr-2 size-4 text-muted-foreground" />
					{m.sidebar_workspace_settings()}
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.Header>

	<Sidebar.Content class={showDesktopPlanner ? 'py-2' : 'px-2 py-3'}>
		{#if showDesktopPlanner}
			<SidebarPlanner onNavigate={navigate} />
		{:else}
			<Sidebar.Group class="p-0">
				<Sidebar.GroupLabel
					class="px-2 text-[11px] tracking-[0.12em] text-sidebar-foreground/48 uppercase"
				>
					{m.sidebar_publish()}
				</Sidebar.GroupLabel>
				<Sidebar.GroupContent>
					<Sidebar.Menu class="gap-1">
						{#each sidebarNavigationItems as item (item.id)}
							<Sidebar.MenuItem>
								<Sidebar.MenuButton
									isActive={isNavigationItemActive(item, currentPath)}
									class="h-10 text-sm"
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
			<div class="pb-1">
				<p
					class="flex h-7 items-center px-2 text-[11px] tracking-[0.1em] text-sidebar-foreground/52 uppercase"
				>
					{m.sidebar_workspace()}
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
								<span>{item.label}</span>
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					{/each}
				</Sidebar.Menu>
			</div>
		{/if}
		<Sidebar.Menu class={showDesktopPlanner ? 'border-t border-sidebar-border pt-1' : ''}>
			<Sidebar.MenuItem>
				<DropdownMenu.Root>
					<DropdownMenu.Trigger>
						{#snippet child({ props })}
							<Sidebar.MenuButton
								{...props}
								size="lg"
								class="data-[state=open]:bg-sidebar-accent"
								tooltipContent={m.sidebar_profile_appearance()}
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
						<AccountPreferencesMenu onNavigate={() => sidebar.setOpenMobile(false)} />
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Sidebar.MenuItem>
		</Sidebar.Menu>
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
