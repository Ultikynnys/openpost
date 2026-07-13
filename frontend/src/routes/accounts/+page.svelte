<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { auth } from '$lib/stores/auth';
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
	import ChevronDownIcon from 'lucide-svelte/icons/chevron-down';
	import { getPlatformName, getPlatformColor } from '$lib/utils';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import UsersIcon from 'lucide-svelte/icons/users';
	import XIcon from 'lucide-svelte/icons/x';
	import { Skeleton } from '$lib/components/ui/skeleton';

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
			'Select workspace'
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

	function showConnectError(value: unknown, fallback = 'Could not start account connection') {
		const message = connectErrorMessage(value, fallback);
		const lower = message.toLowerCase();
		const needsBilling = lower.includes('subscription') || lower.includes('social account limit');
		showToast(
			message,
			needsBilling ? { href: '/settings?tab=billing#billing', label: 'Open billing' } : undefined
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

	function openEditAccount(account: SocialAccount) {
		editingAccount = account;
		editAccountSlug = (account as SocialAccount & { slug?: string }).slug ?? '';
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
			if (err) throw new Error(err.detail || 'Failed to update account slug');
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
			replaceState(window.location.pathname, {});
		}

		const unsubscribe = auth.subscribe(async (state) => {
			if (!state.isLoading && !state.isAuthenticated) {
				goto(resolve('/login'));
			} else if (!state.isLoading && state.isAuthenticated) {
				try {
					const { data, error: err } = await client.GET('/workspaces');
					workspaces = data ?? [];
					if (workspaces && workspaces.length > 0) {
						selectedWorkspaceId = workspaces[0].id;
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
			showToast('Create a workspace before connecting social accounts.');
			return;
		}
		try {
			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: { path: { platform: 'x' }, query: { workspace_id: selectedWorkspaceId } }
			});
			if (err) throw new Error((err as any).detail || 'Failed to get X auth URL');
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			showConnectError(e, 'Failed to get X auth URL');
		}
	}

	async function connectMastodon(options: { serverName?: string; instanceURL?: string }) {
		if (!selectedWorkspaceId) {
			showToast('Create a workspace before connecting social accounts.');
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
			showConnectError(e, 'Failed to get Mastodon auth URL');
		}
	}

	async function connectCustomMastodon() {
		const instanceURL = customMastodonInstance.trim();
		if (!instanceURL) {
			showToast('Enter a Mastodon instance URL.');
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
			showToast('Create a workspace before connecting social accounts.');
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
			blueskyError = 'Please enter both handle and app password';
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
			if (err) throw new Error(err.detail || 'Login failed');
			blueskyModalOpen = false;
			await loadAccounts();
		} catch (e) {
			blueskyError = (e as Error).message;
			showConnectError(e, 'Login failed');
		} finally {
			blueskyLoading = false;
		}
	}

	async function connectLinkedIn() {
		if (!selectedWorkspaceId) {
			showToast('Create a workspace before connecting social accounts.');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'linkedin' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (err) throw new Error((err as any).detail || 'Failed to get LinkedIn auth URL');
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			showConnectError(e, 'Failed to get LinkedIn auth URL');
		}
	}

	async function connectThreads() {
		if (!selectedWorkspaceId) {
			showToast('Create a workspace before connecting social accounts.');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'threads' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (err) throw new Error((err as any).detail || 'Failed to get Threads auth URL');
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			showConnectError(e, 'Failed to get Threads auth URL');
		}
	}

	async function connectTikTok() {
		if (!selectedWorkspaceId) {
			showToast('Create a workspace before connecting social accounts.');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'tiktok' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (err) throw new Error((err as any).detail || 'Failed to get TikTok auth URL');
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			showConnectError(e, 'Failed to get TikTok auth URL');
		}
	}

	async function connectFacebook() {
		if (!selectedWorkspaceId) {
			showToast('Create a workspace before connecting social accounts.');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'facebook' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (err) throw new Error((err as any).detail || 'Failed to get Facebook auth URL');
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			showConnectError(e, 'Failed to get Facebook auth URL');
		}
	}

	async function connectInstagram() {
		if (!selectedWorkspaceId) {
			showToast('Create a workspace before connecting social accounts.');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'instagram' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (err) throw new Error((err as any).detail || 'Failed to get Instagram auth URL');
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			showConnectError(e, 'Failed to get Instagram auth URL');
		}
	}

	async function connectYouTube() {
		if (!selectedWorkspaceId) {
			showToast('Create a workspace before connecting social accounts.');
			return;
		}

		try {
			localStorage.setItem('oauth_workspace_id', selectedWorkspaceId);

			const { data, error: err } = await client.GET('/accounts/{platform}/auth-url', {
				params: {
					path: { platform: 'youtube' },
					query: { workspace_id: selectedWorkspaceId }
				}
			});
			if (err) throw new Error((err as any).detail || 'Failed to get YouTube auth URL');
			if (data?.url) window.location.href = data.url;
		} catch (e) {
			showConnectError(e, 'Failed to get YouTube auth URL');
		}
	}

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
		if (!provider.configured) return 'Not configured';
		if (isCustomMastodonProvider(provider)) {
			return 'Connect any public Mastodon instance';
		}
		if (provider.platform === 'mastodon' && provider.instance_url) {
			return provider.instance_url.replace('https://', '');
		}
		switch (provider.platform) {
			case 'x':
				return 'Post tweets';
			case 'threads':
				return 'Post to Threads';
			case 'bluesky':
				return 'Post to Bluesky';
			case 'linkedin':
				return 'Post to LinkedIn';
			case 'instagram':
				return 'Post to Instagram Business';
			case 'facebook':
				return 'Post to Facebook Pages';
			case 'youtube':
				return 'Upload YouTube videos and Shorts';
			case 'tiktok':
				return 'Post short-form video to TikTok';
			default:
				return 'Connect account';
		}
	}

	function providerStatus(provider: ProviderInfo): string {
		if (provider.status) return provider.status;
		return provider.configured ? 'available' : 'needs_configuration';
	}

	function providerStatusLabel(provider: ProviderInfo): string {
		switch (providerStatus(provider)) {
			case 'available':
				return 'Available';
			case 'planned':
				return 'Planned';
			case 'needs_configuration':
				return 'Needs app config';
			default:
				return provider.configured ? 'Available' : 'Unavailable';
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
		if (providerStatus(provider) === 'planned') return 'Planned';
		return provider.configured ? 'Connect' : 'Unavailable';
	}

	function visibleProviderCapabilities(provider: ProviderInfo): string[] {
		return (provider.capabilities ?? []).slice(0, 4);
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
			showToast('Create a workspace before connecting social accounts.');
			return false;
		}

		const query: { workspace_id: string; server_name?: string; instance_url?: string } = {
			workspace_id: selectedWorkspaceId
		};
		if (isCustomMastodonProvider(provider)) {
			const instanceURL = customMastodonInstance.trim();
			if (!instanceURL) {
				showToast('Enter a Mastodon instance URL.');
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
			showConnectError(e, 'Could not start Mastodon connection');
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

	const accountsByPlatform = $derived.by(() => {
		const grouped = new SvelteMap<string, SocialAccount[]>();
		for (const acc of accounts) {
			const key = acc.platform;
			if (!grouped.has(key)) grouped.set(key, []);
			grouped.get(key)!.push(acc);
		}
		return grouped;
	});
</script>

<svelte:head>
	<title>Connected Accounts - OpenPost</title>
</svelte:head>

{#if toastMessage}
	<div
		class="pointer-events-auto fixed right-4 bottom-4 z-50 mb-4 flex max-w-md items-center gap-3 rounded-lg border bg-background px-4 py-3 shadow-lg"
	>
		<span class="text-sm">{toastMessage}</span>
		{#if toastActionHref && toastActionLabel}
			<Button href={toastActionHref} variant="outline" size="sm">{toastActionLabel}</Button>
		{/if}
		<button onclick={clearToast} class="text-muted-foreground hover:text-foreground">
			<span class="sr-only">Close</span>
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
		<h2 class="mb-2 text-xl font-semibold">No Workspaces Found</h2>
		<p class="mb-4 text-sm text-muted-foreground">
			Create a workspace first before connecting social accounts.
		</p>
		<Button href="/">Create Workspace</Button>
	</div>
{:else}
	<PageContainer
		title="Accounts"
		description="Connect and manage your social accounts."
		icon={UsersIcon}
	>
		{#if error}
			<div
				class="mb-4 flex items-center gap-2 rounded-lg border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive"
			>
				{error}
				<Button variant="ghost" size="sm" onclick={() => (error = '')}>Dismiss</Button>
			</div>
		{/if}

		<!-- Workspace Selector -->
		<div class="mb-6">
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button {...props} variant="outline" class="gap-2">
							<span
								class="flex h-6 w-6 items-center justify-center rounded-md bg-primary/10 text-xs font-bold text-primary"
							>
								{selectedWorkspaceName.slice(0, 2).toUpperCase()}
							</span>
							<span class="truncate">{selectedWorkspaceName}</span>
							<ChevronDownIcon class="size-3.5 opacity-50" />
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content class="w-56 rounded-lg" align="start" side="bottom" sideOffset={6}>
					<DropdownMenu.Label class="text-xs text-muted-foreground">Workspaces</DropdownMenu.Label>
					{#each workspaces as workspace (workspace.id)}
						<DropdownMenu.Item
							onSelect={() => {
								selectedWorkspaceId = workspace.id;
							}}
							class="gap-2 p-2"
						>
							<span
								class="flex size-6 items-center justify-center rounded-md bg-primary/10 text-xs font-bold text-primary"
							>
								{workspace.name.slice(0, 2).toUpperCase()}
							</span>
							<span class="truncate">{workspace.name}</span>
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		</div>

		<!-- Connected Accounts -->
		<div class="mb-8">
			<h2 class="mb-4 text-lg font-semibold">Connected Accounts</h2>

			{#if accountsLoading}
				<div class="space-y-3">
					<Skeleton class="h-12 rounded-lg" />
					<Skeleton class="h-12 rounded-lg" />
					<Skeleton class="h-12 rounded-lg" />
				</div>
			{:else if !accounts || accounts.length === 0}
				<EmptyState
					icon={UsersIcon}
					title="No accounts connected"
					description="Connect a platform below to get started"
					variant="muted"
					size="md"
				/>
			{:else}
				<div class="space-y-3">
					{#each [...accountsByPlatform.entries()] as [platform, platformAccounts] (platform)}
						<div class="rounded-lg border bg-card">
							<div class="flex items-center gap-3 border-b px-4 py-3">
								<div
									class="flex h-9 w-9 items-center justify-center rounded-full {getPlatformColor(
										platform
									)}"
								>
									<PlatformIcon {platform} class="h-4 w-4 text-white" />
								</div>
								<div class="flex-1">
									<h3 class="text-sm font-medium">{getPlatformName(platform)}</h3>
									<p class="text-sm text-muted-foreground">
										{platformAccounts.length} account{platformAccounts.length !== 1 ? 's' : ''}
									</p>
								</div>
							</div>
							<div class="divide-y">
								{#each platformAccounts as account (account.id)}
									<div class="flex items-center justify-between px-4 py-3">
										<div class="flex items-center gap-3">
											<div class="flex h-7 w-7 items-center justify-center rounded-full bg-muted">
												<PlatformIcon platform={account.platform} class="h-3.5 w-3.5" />
											</div>
											<div>
												<p class="text-sm font-medium">
													{accountDisplayName(account)}
												</p>
												<p class="text-sm text-muted-foreground">
													Slug:
													<span class="font-mono"
														>{(account as SocialAccount & { slug?: string }).slug ||
															'not set'}</span
													>
													· {account.is_active ? 'Connected' : 'Disconnected'}
												</p>
											</div>
										</div>
										<div class="flex items-center gap-2">
											<Button
												variant="outline"
												size="sm"
												onclick={() => openEditAccount(account)}
												class="text-xs"
											>
												Edit Slug
											</Button>
											<Button
												variant="ghost"
												size="sm"
												onclick={() => disconnectAccount(account.id)}
												class="text-xs text-muted-foreground hover:text-destructive"
											>
												Disconnect
											</Button>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Connect a Platform -->
		<div>
			<h2 class="mb-4 text-lg font-semibold">Connect a Platform</h2>

			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
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
											onclick={() => openMastodonCode(provider)}>Code</Button
										>
										<Button onclick={() => connectProvider(provider)} size="sm">Connect</Button>
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
							{#if visibleProviderCapabilities(provider).length > 0}
								<div class="mt-3 flex flex-wrap gap-1.5">
									{#each visibleProviderCapabilities(provider) as capability (capability)}
										<span
											class="rounded-full bg-muted px-2 py-0.5 text-[11px] text-muted-foreground"
										>
											{capability}
										</span>
									{/each}
								</div>
							{/if}
							{#if isCustomMastodonProvider(provider)}
								<form
									class="mt-4 grid gap-2 sm:grid-cols-[1fr_auto_auto]"
									onsubmit={(e: SubmitEvent) => {
										e.preventDefault();
										connectCustomMastodon();
									}}
								>
									<div class="space-y-1.5">
										<Label for="custom-mastodon-instance" class="text-xs">Instance URL</Label>
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
										onclick={() => openMastodonCode(provider)}>Code</Button
									>
									<Button type="submit" size="sm" class="self-end" disabled={customMastodonLoading}>
										{customMastodonLoading ? 'Connecting...' : 'Connect'}
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
			<Dialog.Title>Connect Bluesky</Dialog.Title>
			<Dialog.Description>
				Enter your Bluesky handle and an app password. You can create an app password in Bluesky
				Settings &gt; App Passwords.
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
				<Label for="bluesky-handle">Handle</Label>
				<Input
					type="text"
					id="bluesky-handle"
					bind:value={blueskyHandle}
					placeholder="user.bsky.social"
					required
				/>
			</div>
			<div class="space-y-2">
				<Label for="bluesky-password">App Password</Label>
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
					<Button variant="outline" type="button">Cancel</Button>
				</Dialog.Close>
				<Button type="submit" disabled={blueskyLoading}>
					{blueskyLoading ? 'Connecting...' : 'Connect'}
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={editAccountDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Edit Account Slug</Dialog.Title>
			<Dialog.Description>
				Slugs are stable shortcuts for the CLI, for example
				<code class="rounded bg-muted px-1 py-0.5"
					>openpost post create --accounts {editAccountSlug || 'main-x'}</code
				>.
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
				<div class="rounded-md border bg-muted/30 p-3 text-sm">
					<div class="font-medium">{accountDisplayName(editingAccount)}</div>
					<div class="text-muted-foreground">{getPlatformName(editingAccount.platform)}</div>
				</div>
				<div class="space-y-2">
					<Label for="account-slug">Slug</Label>
					<Input
						id="account-slug"
						bind:value={editAccountSlug}
						placeholder="main-x"
						pattern={accountSlugPattern}
						required
					/>
					<p class="text-xs text-muted-foreground">
						Use lowercase letters, numbers, and hyphens. Slugs must be unique within this workspace.
					</p>
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
						<Button variant="outline" type="button">Cancel</Button>
					</Dialog.Close>
					<Button type="submit" disabled={editAccountLoading || !editAccountSlug.trim()}>
						{editAccountLoading ? 'Saving...' : 'Save Slug'}
					</Button>
				</div>
			</form>
		{/if}
	</Dialog.Content>
</Dialog.Root>
