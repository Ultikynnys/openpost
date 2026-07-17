<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { client, type Workspace, type SocialAccount, type ProviderInfo } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { goto, replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageContainer from '$lib/components/page-container.svelte';
	import EmptyState from '$lib/components/empty-state.svelte';
	import MoreHorizontalIcon from 'lucide-svelte/icons/ellipsis';
	import { getPlatformName, getPlatformColor } from '$lib/utils';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import UsersIcon from 'lucide-svelte/icons/users';
	import XIcon from 'lucide-svelte/icons/x';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { m } from '$lib/paraglide/messages';

	let workspaces = $state<Workspace[] | null>(null);
	let selectedWorkspaceId = $state('');
	let loading = $state(true);
	let error = $state('');

	let accounts = $state<SocialAccount[]>([]);
	let accountsLoading = $state(false);

	let providerEntries = $state.raw<ProviderInfo[]>([]);
	let providersLoading = $state(false);
	let customMastodonInstance = $state('');
	let customMastodonLoading = $state(false);
	let selectedWorkspaceName = $derived(
		workspaces?.find((workspace) => workspace.id === selectedWorkspaceId)?.name ||
			m.accounts_select_workspace()
	);
	let toastMessage = $state('');
	let toastActionHref = $state('');
	let toastActionLabel = $state('');

	let blueskyModalOpen = $state(false);
	let blueskyHandle = $state('');
	let blueskyAppPassword = $state('');
	let blueskyLoading = $state(false);
	let blueskyError = $state('');

	let editAccountDialogOpen = $state(false);
	let editingAccount = $state<SocialAccount | null>(null);
	let editAccountSlug = $state('');
	let editAccountLoading = $state(false);
	let editAccountError = $state('');
	const accountSlugPattern = '[a-z0-9][a-z0-9-]{0,62}';

	function clearToast() {
		toastMessage = '';
		toastActionHref = '';
		toastActionLabel = '';
	}

	function showToast(message: string, action?: { href: string; label: string }) {
		error = '';
		toastMessage = message;
		toastActionHref = action?.href ?? '';
		toastActionLabel = action?.label ?? '';
	}

	function connectErrorMessage(value: unknown, fallback: string): string {
		if (value && typeof value === 'object') {
			const maybeError = value as { detail?: string; message?: string };
			return maybeError.detail || maybeError.message || fallback;
		}
		return fallback;
	}

	function showConnectError(value: unknown, fallback: string = m.accounts_connect_failed()) {
		const message = connectErrorMessage(value, fallback);
		const lower = message.toLowerCase();
		const needsBilling = lower.includes('subscription') || lower.includes('social account limit');
		showToast(
			message,
			needsBilling
				? { href: '/settings?tab=billing#billing', label: m.accounts_open_billing() }
				: undefined
		);
	}

	async function loadAccounts() {
		if (!selectedWorkspaceId) return;
		accountsLoading = true;
		try {
			const { data, error: err } = await client.GET('/accounts', {
				params: { query: { workspace_id: selectedWorkspaceId } }
			});
			accounts = data ?? [];
		} catch (e) {
			console.error('Failed to load accounts:', e);
			accounts = [];
		} finally {
			accountsLoading = false;
		}
	}

	async function loadProviders() {
		providersLoading = true;
		try {
			const { data, error: err } = await client.GET('/accounts/providers');
			if (err) throw new Error(err.detail ?? 'Failed to load account providers');
			providerEntries = data ?? [];
		} catch (e) {
			console.error('Failed to load account providers:', e);
			providerEntries = [];
		} finally {
			providersLoading = false;
		}
	}

	async function disconnectAccount(accountId: string) {
		try {
			await client.DELETE('/accounts/{account_id}', {
				params: { path: { account_id: accountId } }
			});
			await loadAccounts();
		} catch (e) {
			error = (e as Error).message;
		}
	}

	function accountDisplayName(account: SocialAccount): string {
		if (account.account_username) return `@${account.account_username}`;
		if (account.instance_url) return account.instance_url.replace('https://', '');
		return account.account_id || account.platform;
	}

	function accountSlug(account: SocialAccount): string {
		return account.slug || account.account_username || account.account_id || account.platform;
	}

	function accountServer(account: SocialAccount): string {
		if (account.platform !== 'mastodon' || !account.instance_url) return '';
		try {
			return new URL(account.instance_url).host;
		} catch {
			return account.instance_url.replace(/^https?:\/\//, '').replace(/\/$/, '');
		}
	}

	function openEditAccount(account: SocialAccount) {
		editingAccount = account;
		editAccountSlug = account.slug ?? '';
		editAccountError = '';
		editAccountDialogOpen = true;
	}

	async function updateAccountSlug() {
		if (!editingAccount) return;
		editAccountLoading = true;
		editAccountError = '';
		try {
			const { error: err } = await client.PATCH('/accounts/{account_id}', {
				params: { path: { account_id: editingAccount.id } },
				body: { slug: editAccountSlug.trim() }
			});
			if (err) throw new Error(err.detail || m.accounts_update_slug_failed());
			editAccountDialogOpen = false;
			editingAccount = null;
			await loadAccounts();
		} catch (e) {
			editAccountError = (e as Error).message;
		} finally {
			editAccountLoading = false;
		}
	}

	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		const urlError = params.get('error');
		if (urlError) {
			error = urlError;
			replaceState(resolve(window.location.pathname as '/'), {});
		}

		const unsubscribe = auth.subscribe(async (state) => {
			if (!state.isLoading && !state.isAuthenticated) {
				goto(resolve('/login'));
			} else if (!state.isLoading && state.isAuthenticated) {
				try {
					if (workspaceCtx.workspaces.length === 0) {
						await workspaceCtx.initialize();
					}
					const { data, error: err } = await client.GET('/workspaces');
					if (err) throw new Error(err.detail ?? 'Failed to load workspaces');
					workspaces = data ?? [];
					if (workspaces && workspaces.length > 0) {
						const currentWorkspace = workspaces.find(
							(workspace) => workspace.id === workspaceCtx.currentWorkspace?.id
						);
						selectedWorkspaceId = (currentWorkspace ?? workspaces[0]).id;
						await loadAccounts();
					}
					await loadProviders();
				} catch (e) {
					console.error('Failed to load workspaces:', e);
				} finally {
					loading = false;
				}
			}
		});
		return unsubscribe;
	});

	$effect(() => {
		if (selectedWorkspaceId) {
			loadAccounts();
		} else {
			accounts = [];
		}
	});

	async function connectTwitter() {
		if (!selectedWorkspaceId) {
			showToast(m.accounts_create_workspace_first());
			return;
		}
		try {
			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: { path: { platform: 'x' }, query: { workspace_id: selectedWorkspaceId } }
			});
			if (err) throw new Error((err as any).detail || 'Failed to get X auth URL');
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			showConnectError(e);
		}
	}

	async function connectMastodon(options: { serverName?: string; instanceURL?: string }) {
		if (!selectedWorkspaceId) {
			showToast(m.accounts_create_workspace_first());
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);
			if (options.instanceURL) {
				localStorage.setItem('oauth_mastodon_instance_url', options.instanceURL);
				localStorage.removeItem('oauth_mastodon_server');
			} else if (options.serverName) {
				localStorage.setItem('oauth_mastodon_server', options.serverName);
				localStorage.removeItem('oauth_mastodon_instance_url');
			}

			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'mastodon' },
					query: {
						workspace_id: selectedWorkspaceId,
						server_name: options.serverName,
						instance_url: options.instanceURL
					}
				}
			});
			if (err) throw new Error((err as any).detail || 'Failed to get Mastodon auth URL');
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			showConnectError(e);
		}
	}

	async function connectCustomMastodon() {
		const instanceURL = customMastodonInstance.trim();
		if (!instanceURL) {
			showToast(m.accounts_enter_mastodon_instance());
			return;
		}
		customMastodonLoading = true;
		try {
			await connectMastodon({ instanceURL });
		} finally {
			customMastodonLoading = false;
		}
	}

	async function connectBluesky() {
		if (!selectedWorkspaceId) {
			showToast(m.accounts_create_workspace_first());
			return;
		}
		clearToast();
		blueskyHandle = '';
		blueskyAppPassword = '';
		blueskyError = '';
		blueskyModalOpen = true;
	}

	async function submitBlueskyLogin() {
		if (!blueskyHandle.trim() || !blueskyAppPassword.trim()) {
			blueskyError = m.accounts_bluesky_fields_required();
			return;
		}

		blueskyLoading = true;
		blueskyError = '';

		try {
			const { error: err } = await client.POST('/accounts/bluesky/login', {
				body: {
					workspace_id: selectedWorkspaceId,
					handle: blueskyHandle.trim(),
					app_password: blueskyAppPassword.trim()
				}
			});
			if (err) throw new Error(err.detail || m.accounts_login_failed());
			blueskyModalOpen = false;
			await loadAccounts();
		} catch (e) {
			blueskyError = (e as Error).message;
			showConnectError(e, m.accounts_login_failed());
		} finally {
			blueskyLoading = false;
		}
	}

	async function connectOAuthProvider(platform: string) {
		if (!selectedWorkspaceId) {
			showToast(m.accounts_create_workspace_first());
			return;
		}
		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);
			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (err) throw new Error((err as any).detail || m.accounts_connect_failed());
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			showConnectError(e);
		}
	}

	const connectLinkedIn = () => connectOAuthProvider('linkedin');
	const connectThreads = () => connectOAuthProvider('threads');
	const connectTikTok = () => connectOAuthProvider('tiktok');
	const connectFacebook = () => connectOAuthProvider('facebook');
	const connectInstagram = () => connectOAuthProvider('instagram');
	const connectYouTube = () => connectOAuthProvider('youtube');

	function providerKey(provider: ProviderInfo): string {
		if (provider.platform === 'mastodon') {
			return provider.instance_url || provider.name || provider.platform;
		}
		return provider.platform;
	}

	function providerTitle(provider: ProviderInfo): string {
		if (provider.platform === 'mastodon' && provider.name) return provider.name;
		return provider.display_name || getPlatformName(provider.platform);
	}

	function providerDescription(provider: ProviderInfo): string {
		if (provider.description) return provider.description;
		if (!provider.configured) return m.accounts_provider_admin_enable();
		if (isCustomMastodonProvider(provider)) {
			return m.accounts_provider_custom_mastodon();
		}
		if (provider.platform === 'mastodon' && provider.instance_url) {
			return provider.instance_url.replace('https://', '');
		}
		switch (provider.platform) {
			case 'x':
				return m.accounts_provider_x();
			case 'threads':
				return m.accounts_provider_threads();
			case 'bluesky':
				return m.accounts_provider_bluesky();
			case 'linkedin':
				return m.accounts_provider_linkedin();
			case 'instagram':
				return m.accounts_provider_instagram();
			case 'facebook':
				return m.accounts_provider_facebook();
			case 'youtube':
				return m.accounts_provider_youtube();
			case 'tiktok':
				return m.accounts_provider_tiktok();
			default:
				return m.accounts_provider_default();
		}
	}

	function providerStatus(provider: ProviderInfo): string {
		if (provider.status) return provider.status;
		return provider.configured ? 'available' : 'needs_configuration';
	}

	function providerStatusLabel(provider: ProviderInfo): string {
		switch (providerStatus(provider)) {
			case 'available':
				return m.accounts_provider_available();
			case 'planned':
				return m.accounts_provider_planned();
			case 'needs_configuration':
				return m.accounts_provider_admin_required();
			default:
				return provider.configured
					? m.accounts_provider_available()
					: m.accounts_provider_unavailable();
		}
	}

	function providerStatusClass(provider: ProviderInfo): string {
		switch (providerStatus(provider)) {
			case 'available':
				return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300';
			case 'planned':
				return 'border-blue-500/20 bg-blue-500/10 text-blue-700 dark:text-blue-300';
			case 'needs_configuration':
				return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-300';
			default:
				return 'border-muted bg-muted text-muted-foreground';
		}
	}

	function providerCanConnect(provider: ProviderInfo): boolean {
		return provider.configured && providerStatus(provider) !== 'planned';
	}

	function providerActionLabel(provider: ProviderInfo): string {
		if (providerStatus(provider) === 'planned') return m.accounts_provider_planned();
		return provider.configured ? m.common_connect() : m.accounts_provider_ask_admin();
	}

	function isCustomMastodonProvider(provider: ProviderInfo): boolean {
		return provider.platform === 'mastodon' && provider.configured && !provider.instance_url;
	}

	function rememberMastodonProvider(provider: ProviderInfo) {
		if (!selectedWorkspaceId) return;
		localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);
		if (isCustomMastodonProvider(provider)) {
			const instanceURL = customMastodonInstance.trim();
			if (instanceURL) {
				localStorage.setItem('oauth_mastodon_instance_url', instanceURL);
				localStorage.removeItem('oauth_mastodon_server');
			}
			return;
		}
		const serverName = provider.name || provider.instance_url || '';
		if (serverName) {
			localStorage.setItem('oauth_mastodon_server', serverName);
			localStorage.removeItem('oauth_mastodon_instance_url');
		}
	}

	async function canOpenMastodonCode(provider: ProviderInfo): Promise<boolean> {
		if (!selectedWorkspaceId) {
			showToast(m.accounts_create_workspace_first());
			return false;
		}

		const query: { workspace_id: string; server_name?: string; instance_url?: string } = {
			workspace_id: selectedWorkspaceId
		};
		if (isCustomMastodonProvider(provider)) {
			const instanceURL = customMastodonInstance.trim();
			if (!instanceURL) {
				showToast(m.accounts_enter_mastodon_instance());
				return false;
			}
			query.instance_url = instanceURL;
		} else {
			query.server_name = provider.name || provider.instance_url || '';
		}

		try {
			const { error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: { path: { platform: 'mastodon' }, query }
			});
			if (err) throw new Error((err as any).detail || 'Could not start Mastodon connection');
			return true;
		} catch (e) {
			showConnectError(e);
			return false;
		}
	}

	async function openMastodonCode(provider: ProviderInfo) {
		if (!(await canOpenMastodonCode(provider))) return;
		rememberMastodonProvider(provider);
		goto(resolve('/accounts/mastodon/callback'));
	}

	function connectProvider(provider: ProviderInfo) {
		if (!providerCanConnect(provider)) return;
		switch (provider.platform) {
			case 'x':
				connectTwitter();
				break;
			case 'mastodon':
				if (isCustomMastodonProvider(provider)) {
					connectCustomMastodon();
				} else {
					connectMastodon({ serverName: provider.name || provider.instance_url || '' });
				}
				break;
			case 'threads':
				connectThreads();
				break;
			case 'bluesky':
				connectBluesky();
				break;
			case 'linkedin':
				connectLinkedIn();
				break;
			case 'instagram':
				connectInstagram();
				break;
			case 'facebook':
				connectFacebook();
				break;
			case 'youtube':
				connectYouTube();
				break;
			case 'tiktok':
				connectTikTok();
				break;
		}
	}
</script>

<svelte:head>
	<title>{m.accounts_page_title()}</title>
</svelte:head>

{#if toastMessage}
	<div
		class="pointer-events-auto fixed right-4 bottom-4 z-50 mb-4 flex max-w-md items-center gap-3 rounded-lg border bg-background px-4 py-3 shadow-lg"
	>
		<span class="text-sm" role="status" aria-live="polite">{toastMessage}</span>
		{#if toastActionHref && toastActionLabel}
			<Button href={toastActionHref} variant="outline" size="sm">{toastActionLabel}</Button>
		{/if}
		<button
			onclick={clearToast}
			class="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
			aria-label={m.common_dismiss()}
		>
			<XIcon class="size-4" />
		</button>
	</div>
{/if}

{#if loading}
	<div class="mx-auto w-full max-w-6xl px-4 py-6 lg:px-8">
		<div class="mb-6 flex items-center gap-2">
			<Skeleton class="h-8 w-8 rounded-md" />
			<Skeleton class="h-7 w-48" />
		</div>
		<div class="mb-6"><Skeleton class="h-9 w-40 rounded-md" /></div>
		<div class="mb-8 space-y-3">
			<Skeleton class="h-6 w-32" />
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
				<Skeleton class="h-28 rounded-lg" />
				<Skeleton class="h-28 rounded-lg" />
			</div>
		</div>
		<div class="space-y-3">
			<Skeleton class="h-6 w-40" />
			<Skeleton class="h-40 rounded-lg" />
		</div>
	</div>
{:else if !workspaces || workspaces.length === 0}
	<div class="mx-auto max-w-md px-4 py-16 text-center">
		<h2 class="mb-2 text-xl font-semibold">{m.accounts_no_workspaces_title()}</h2>
		<p class="mb-4 text-sm text-muted-foreground">
			{m.accounts_no_workspaces_body()}
		</p>
		<Button href="/">{m.accounts_create_workspace()}</Button>
	</div>
{:else}
	<PageContainer
		title={m.accounts_heading()}
		description={m.accounts_description()}
		icon={UsersIcon}
	>
		{#if error}
			<div
				class="mb-4 flex items-center gap-2 rounded-lg border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
			>
				{error}
				<Button variant="ghost" size="sm" onclick={() => (error = '')}>{m.common_dismiss()}</Button>
			</div>
		{/if}

		<!-- Connected Accounts -->
		<div class="mb-10">
			<div class="mb-4 flex flex-wrap items-end justify-between gap-3">
				<div>
					<h2 class="text-base font-semibold">{m.accounts_connected_channels()}</h2>
					<p class="mt-1 text-sm text-muted-foreground">
						{m.accounts_connection_summary({
							count: accounts.length,
							workspace: selectedWorkspaceName
						})}
					</p>
				</div>
				<Button href="/" size="sm">{m.accounts_create_post()}</Button>
			</div>

			{#if accountsLoading}
				<div class="space-y-3">
					<Skeleton class="h-12 rounded-lg" />
					<Skeleton class="h-12 rounded-lg" />
					<Skeleton class="h-12 rounded-lg" />
				</div>
			{:else if !accounts || accounts.length === 0}
				<EmptyState
					icon={UsersIcon}
					title={m.accounts_empty_title()}
					description={m.accounts_empty_body()}
					variant="muted"
					size="md"
				/>
			{:else}
				<div
					class="grid gap-px overflow-hidden rounded-lg border bg-border sm:grid-cols-2 xl:grid-cols-3"
				>
					{#each accounts as account (account.id)}
						<article class="flex min-h-28 flex-col justify-between gap-3 bg-background p-4">
							<div class="flex items-start gap-3">
								<div
									class="flex size-10 shrink-0 items-center justify-center rounded-lg {getPlatformColor(
										account.platform
									)}"
								>
									<PlatformIcon platform={account.platform} class="size-5 text-white" />
								</div>
								<div class="min-w-0 flex-1">
									<div class="flex items-center gap-2">
										<h3 class="truncate text-sm font-semibold">
											{getPlatformName(account.platform)}
										</h3>
										{#if !account.is_active}
											<span
												class="size-1.5 rounded-full bg-amber-500"
												aria-label={m.accounts_connection_paused()}
											></span>
										{/if}
									</div>
									<p class="mt-1 truncate text-sm text-muted-foreground">
										{accountDisplayName(account)}
									</p>
								</div>
								<DropdownMenu.Root>
									<DropdownMenu.Trigger>
										{#snippet child({ props })}
											<Button
												{...props}
												variant="ghost"
												size="icon-sm"
												aria-label={m.accounts_actions_for({
													account: accountDisplayName(account)
												})}
											>
												<MoreHorizontalIcon class="size-4" />
											</Button>
										{/snippet}
									</DropdownMenu.Trigger>
									<DropdownMenu.Content align="end" class="w-48">
										<DropdownMenu.Item onclick={() => openEditAccount(account)}
											>{m.accounts_details()}</DropdownMenu.Item
										>
										<DropdownMenu.Separator />
										<DropdownMenu.Item
											class="text-destructive"
											onclick={() => disconnectAccount(account.id)}
											>{m.common_disconnect()}</DropdownMenu.Item
										>
									</DropdownMenu.Content>
								</DropdownMenu.Root>
							</div>
							<div
								class="flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground"
							>
								<span class="inline-flex min-w-0 items-center gap-1.5">
									<span>{m.accounts_shortcut()}</span>
									<code
										class="max-w-44 truncate rounded bg-muted px-1.5 py-0.5 font-mono text-[0.6875rem] text-foreground"
										>{accountSlug(account)}</code
									>
								</span>
								{#if accountServer(account)}
									<span class="truncate">{m.accounts_server()}: {accountServer(account)}</span>
								{/if}
								{#if !account.is_active}
									<span class="text-amber-700 dark:text-amber-300"
										>{m.accounts_connection_paused()}</span
									>
								{/if}
							</div>
						</article>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Connect a Platform -->
		<div>
			<h2 class="mb-1 text-base font-semibold">{m.accounts_add_channel()}</h2>
			<p class="mb-4 text-sm text-muted-foreground">
				{m.accounts_add_channel_body()}
			</p>

			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
				{#if providersLoading}
					<Skeleton class="h-20 rounded-lg" />
					<Skeleton class="h-20 rounded-lg" />
					<Skeleton class="h-20 rounded-lg" />
					<Skeleton class="h-20 rounded-lg" />
				{:else}
					{#each providerEntries as provider (providerKey(provider))}
						<div
							data-testid={`provider-card-${provider.platform}`}
							class="group rounded-lg border bg-card p-4 transition-all hover:shadow-sm {providerCanConnect(
								provider
							)
								? ''
								: 'bg-muted/20'}"
						>
							<div class="flex items-center gap-3">
								<div
									class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full {getPlatformColor(
										provider.platform
									)}"
								>
									<PlatformIcon platform={provider.platform} class="h-4 w-4 text-white" />
								</div>
								<div class="min-w-0 flex-1">
									<div class="flex flex-wrap items-center gap-2">
										<h3 class="text-sm font-medium">{providerTitle(provider)}</h3>
										<span
											class="inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium {providerStatusClass(
												provider
											)}"
										>
											{providerStatusLabel(provider)}
										</span>
									</div>
									<p class="truncate text-sm text-muted-foreground">
										{providerDescription(provider)}
									</p>
								</div>
								{#if provider.platform === 'mastodon' && provider.configured && !isCustomMastodonProvider(provider)}
									<div class="flex gap-1.5">
										<Button
											variant="outline"
											size="sm"
											class="text-xs"
											onclick={() => openMastodonCode(provider)}>{m.accounts_code()}</Button
										>
										<Button onclick={() => connectProvider(provider)} size="sm"
											>{m.common_connect()}</Button
										>
									</div>
								{:else if !isCustomMastodonProvider(provider)}
									<Button
										onclick={() => connectProvider(provider)}
										size="sm"
										disabled={!providerCanConnect(provider)}
									>
										{providerActionLabel(provider)}
									</Button>
								{/if}
							</div>
							{#if isCustomMastodonProvider(provider)}
								<form
									class="mt-4 grid gap-2 sm:grid-cols-[1fr_auto_auto]"
									onsubmit={(e: SubmitEvent) => {
										e.preventDefault();
										connectCustomMastodon();
									}}
								>
									<div class="space-y-1.5">
										<Label for="custom-mastodon-instance" class="text-xs"
											>{m.accounts_instance_url()}</Label
										>
										<Input
											id="custom-mastodon-instance"
											bind:value={customMastodonInstance}
											placeholder="mastodon.social"
											autocomplete="url"
										/>
									</div>
									<Button
										variant="outline"
										size="sm"
										class="self-end"
										onclick={() => openMastodonCode(provider)}>{m.accounts_code()}</Button
									>
									<Button type="submit" size="sm" class="self-end" disabled={customMastodonLoading}>
										{customMastodonLoading ? m.common_connecting() : m.common_connect()}
									</Button>
								</form>
							{/if}
						</div>
					{/each}
				{/if}
			</div>
		</div>
	</PageContainer>
{/if}

<Dialog.Root bind:open={blueskyModalOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>{m.accounts_connect_bluesky()}</Dialog.Title>
			<Dialog.Description>
				{m.accounts_bluesky_description()}
			</Dialog.Description>
		</Dialog.Header>
		<form
			class="space-y-4"
			onsubmit={(e) => {
				e.preventDefault();
				submitBlueskyLogin();
			}}
		>
			<div class="space-y-2">
				<Label for="bluesky-handle">{m.accounts_handle()}</Label>
				<Input
					type="text"
					id="bluesky-handle"
					bind:value={blueskyHandle}
					placeholder="user.bsky.social"
					required
				/>
			</div>
			<div class="space-y-2">
				<Label for="bluesky-password">{m.accounts_app_password()}</Label>
				<Input
					type="password"
					id="bluesky-password"
					bind:value={blueskyAppPassword}
					placeholder="xxxx-xxxx-xxxx-xxxx"
					required
				/>
			</div>
			{#if blueskyError}
				<div
					class="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
				>
					{blueskyError}
				</div>
			{/if}
			<div class="flex justify-end gap-2">
				<Dialog.Close>
					<Button variant="outline" type="button">{m.common_cancel()}</Button>
				</Dialog.Close>
				<Button type="submit" disabled={blueskyLoading}>
					{blueskyLoading ? m.common_connecting() : m.common_connect()}
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={editAccountDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>{m.accounts_details()}</Dialog.Title>
			<Dialog.Description>
				{m.accounts_details_description()}
			</Dialog.Description>
		</Dialog.Header>
		{#if editingAccount}
			<form
				class="space-y-4"
				onsubmit={(e) => {
					e.preventDefault();
					updateAccountSlug();
				}}
			>
				<div class="rounded-md bg-muted/40 p-3 text-sm">
					<div class="font-medium">{accountDisplayName(editingAccount)}</div>
					<div class="text-muted-foreground">{getPlatformName(editingAccount.platform)}</div>
					{#if accountServer(editingAccount)}
						<div class="mt-1 text-xs text-muted-foreground">
							{m.accounts_server()}: {accountServer(editingAccount)}
						</div>
					{/if}
				</div>
				<div class="space-y-3 rounded-md border p-3">
					<h3 class="text-sm font-medium">{m.accounts_developer_shortcut()}</h3>
					<p class="text-xs text-muted-foreground">
						{m.accounts_shortcut_example()}
						<code class="rounded bg-muted px-1 py-0.5"
							>openpost post create --accounts {editAccountSlug || 'main-x'}</code
						>.
					</p>
					<div class="space-y-2">
						<Label for="account-slug">{m.accounts_shortcut()}</Label>
						<Input
							id="account-slug"
							bind:value={editAccountSlug}
							placeholder="main-x"
							pattern={accountSlugPattern}
							required
						/>
						<p class="text-xs text-muted-foreground">
							{m.accounts_shortcut_hint()}
						</p>
					</div>
				</div>
				{#if editAccountError}
					<div
						class="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
					>
						{editAccountError}
					</div>
				{/if}
				<div class="flex justify-end gap-2">
					<Dialog.Close>
						<Button variant="outline" type="button">{m.common_cancel()}</Button>
					</Dialog.Close>
					<Button type="submit" disabled={editAccountLoading || !editAccountSlug.trim()}>
						{editAccountLoading ? m.common_saving() : m.accounts_save_details()}
					</Button>
				</div>
			</form>
		{/if}
	</Dialog.Content>
</Dialog.Root>
