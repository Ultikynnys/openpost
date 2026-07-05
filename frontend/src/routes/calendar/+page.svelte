<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { client, type Post, type SocialAccount, type Workspace } from '$lib/api/client';
	import type { components } from '$lib/api/types';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { getLocaleTag } from '$lib/i18n';
	import { m } from '$lib/paraglide/messages';
	import { ui } from '$lib/stores/ui.svelte';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { cn } from '$lib/utils';
	import CalendarDaysIcon from 'lucide-svelte/icons/calendar-days';
	import CheckIcon from 'lucide-svelte/icons/check';
	import ChevronLeftIcon from 'lucide-svelte/icons/chevron-left';
	import ChevronRightIcon from 'lucide-svelte/icons/chevron-right';
	import ClockIcon from 'lucide-svelte/icons/clock';
	import LayersIcon from 'lucide-svelte/icons/layers';
	import Loader2Icon from 'lucide-svelte/icons/loader-2';
	import RefreshCwIcon from 'lucide-svelte/icons/refresh-cw';

	type Publication = components['schemas']['PublicationResponse'];
	type SetResponse = components['schemas']['SetResponse'];
	type PostDestination = components['schemas']['PostDestinationResponse'];
	type Rendition = components['schemas']['RenditionResponse'];

	type CalendarDay = {
		date: Date;
		key: string;
		outsideMonth: boolean;
		today: boolean;
	};

	type AccountBadge = {
		id: string;
		platform: string;
		label: string;
	};

	type CalendarItem = {
		id: string;
		key: string;
		kind: 'post' | 'publication';
		title: string;
		preview: string;
		status: string;
		scheduledAt: string;
		randomDelayMinutes: number;
		workspaceId: string;
		workspaceName: string;
		accounts: AccountBadge[];
		platforms: string[];
		setName: string;
		mediaCount: number;
		profile: string;
		publication?: Publication;
	};

	let currentMonth = $state(startOfMonth(new Date()));
	let selectedWorkspaceIds = $state<string[]>([]);
	let selectedPlatform = $state('all');
	let posts = $state<Post[]>([]);
	let publications = $state<Publication[]>([]);
	let accountsByWorkspace = $state<Record<string, SocialAccount[]>>({});
	let setsByWorkspace = $state<Record<string, SetResponse[]>>({});
	let loading = $state(true);
	let errorMessage = $state('');
	let successMessage = $state('');
	let draggingKey = $state('');
	let dropTargetKey = $state('');
	let reschedulingKey = $state('');
	let activeRequest = 0;

	const workspaces = $derived(workspaceCtx.workspaces);
	const activeWorkspaceIds = $derived.by(() => {
		if (selectedWorkspaceIds.length > 0) return selectedWorkspaceIds;
		return workspaces.map((workspace) => workspace.id);
	});
	const loadKey = $derived(
		`${monthKey(currentMonth)}|${activeWorkspaceIds.join(',')}|${workspaces.map((w) => w.id).join(',')}`
	);
	const days = $derived.by(() =>
		buildCalendarDays(currentMonth, workspaceCtx.weekStartsOn, new Date())
	);
	const weekdayLabels = $derived.by(() =>
		days.slice(0, 7).map((day) => day.date.toLocaleDateString(getLocaleTag(), { weekday: 'short' }))
	);
	const visibleRange = $derived.by(() => {
		const start = new Date(days[0]?.date ?? currentMonth);
		start.setHours(0, 0, 0, 0);
		const end = new Date(days[days.length - 1]?.date ?? currentMonth);
		end.setHours(23, 59, 59, 999);
		return { start, end };
	});
	const allItems = $derived.by((): CalendarItem[] => {
		const items: CalendarItem[] = [];
		for (const post of posts) {
			const item = postToCalendarItem(post);
			if (item) items.push(item);
		}
		for (const publication of publications) {
			const item = publicationToCalendarItem(publication);
			if (item) items.push(item);
		}
		return items.sort(
			(a, b) =>
				new Date(a.scheduledAt).getTime() - new Date(b.scheduledAt).getTime() ||
				a.title.localeCompare(b.title)
		);
	});
	const availablePlatforms = $derived.by(() => {
		const platforms = new Set<string>();
		for (const item of allItems) {
			for (const platform of item.platforms) platforms.add(platform);
		}
		return Array.from(platforms).sort((a, b) => platformLabel(a).localeCompare(platformLabel(b)));
	});
	const visibleItems = $derived.by(() =>
		allItems.filter((item) => {
			const scheduled = new Date(item.scheduledAt);
			const inVisibleMonth = scheduled >= visibleRange.start && scheduled <= visibleRange.end;
			const platformMatches =
				selectedPlatform === 'all' || item.platforms.includes(selectedPlatform);
			return inVisibleMonth && platformMatches;
		})
	);
	const itemsByDay = $derived.by(() => {
		const map = new Map<string, CalendarItem[]>();
		for (const item of visibleItems) {
			const key = dateKey(new Date(item.scheduledAt));
			const existing = map.get(key) ?? [];
			existing.push(item);
			map.set(key, existing);
		}
		return map;
	});
	const monthItems = $derived(
		visibleItems.filter((item) => monthKey(new Date(item.scheduledAt)) === monthKey(currentMonth))
	);
	const postCount = $derived(monthItems.filter((item) => item.kind === 'post').length);
	const publicationCount = $derived(
		monthItems.filter((item) => item.kind === 'publication').length
	);
	const selectedWorkspaceLabel = $derived.by(() => {
		if (selectedWorkspaceIds.length === 0 || selectedWorkspaceIds.length === workspaces.length) {
			return m.calendar_all_workspaces();
		}
		if (selectedWorkspaceIds.length === 1) {
			return workspaceName(selectedWorkspaceIds[0]);
		}
		return m.calendar_workspace_count({ count: selectedWorkspaceIds.length });
	});

	$effect(() => {
		if (selectedPlatform !== 'all' && !availablePlatforms.includes(selectedPlatform)) {
			selectedPlatform = 'all';
		}
	});

	$effect(() => {
		const validWorkspaceIds = new Set(workspaces.map((workspace) => workspace.id));
		if (selectedWorkspaceIds.some((workspaceId) => !validWorkspaceIds.has(workspaceId))) {
			selectedWorkspaceIds = selectedWorkspaceIds.filter((workspaceId) =>
				validWorkspaceIds.has(workspaceId)
			);
		}
	});

	$effect(() => {
		const key = loadKey;
		void loadCalendarData(key);
	});

	async function loadCalendarData(_key: string) {
		const request = ++activeRequest;
		loading = true;
		errorMessage = '';
		try {
			if (workspaceCtx.workspaces.length === 0 && !workspaceCtx.loading) {
				await workspaceCtx.initialize();
			}
			const workspaceIds =
				selectedWorkspaceIds.length > 0
					? selectedWorkspaceIds
					: workspaceCtx.workspaces.map((workspace) => workspace.id);
			if (workspaceIds.length === 0) {
				if (request !== activeRequest) return;
				posts = [];
				publications = [];
				accountsByWorkspace = {};
				setsByWorkspace = {};
				return;
			}

			const [postGroups, publicationGroups, accountEntries, setEntries] = await Promise.all([
				Promise.all(workspaceIds.map(fetchScheduledPosts)),
				Promise.all(workspaceIds.map(fetchScheduledPublications)),
				Promise.all(workspaceIds.map(fetchAccounts)),
				Promise.all(workspaceIds.map(fetchSets))
			]);

			if (request !== activeRequest) return;
			posts = postGroups.flat();
			publications = publicationGroups.flat();
			accountsByWorkspace = Object.fromEntries(accountEntries);
			setsByWorkspace = Object.fromEntries(setEntries);
		} catch (error) {
			if (request !== activeRequest) return;
			errorMessage = problemMessage(error, m.calendar_failed_load());
			posts = [];
			publications = [];
		} finally {
			if (request === activeRequest) loading = false;
		}
	}

	async function fetchScheduledPosts(workspaceId: string) {
		const out: Post[] = [];
		let offset = 0;
		while (true) {
			const { data, error, response } = await client.GET('/posts', {
				params: { query: { workspace_id: workspaceId, status: 'scheduled', limit: 200, offset } }
			});
			if (error) throw new Error(problemMessage(error, m.calendar_failed_load()));
			out.push(...(data ?? []));
			const hasMore = response.headers.get('X-Has-More') === 'true';
			if (!hasMore) break;
			const nextOffset = Number(response.headers.get('X-Next-Offset') ?? offset + 200);
			if (!Number.isFinite(nextOffset) || nextOffset <= offset) break;
			offset = nextOffset;
		}
		return out;
	}

	async function fetchScheduledPublications(workspaceId: string) {
		const out: Publication[] = [];
		let offset = 0;
		while (true) {
			const { data, error, response } = await client.GET('/publications', {
				params: { query: { workspace_id: workspaceId, status: 'scheduled', limit: 200, offset } }
			});
			if (error) throw new Error(problemMessage(error, m.calendar_failed_load()));
			out.push(...(data ?? []));
			const hasMore = response.headers.get('X-Has-More') === 'true';
			if (!hasMore) break;
			const nextOffset = Number(response.headers.get('X-Next-Offset') ?? offset + 200);
			if (!Number.isFinite(nextOffset) || nextOffset <= offset) break;
			offset = nextOffset;
		}
		return out;
	}

	async function fetchAccounts(workspaceId: string): Promise<[string, SocialAccount[]]> {
		const { data, error } = await client.GET('/accounts', {
			params: { query: { workspace_id: workspaceId } }
		});
		if (error) throw new Error(problemMessage(error, m.calendar_failed_load()));
		return [workspaceId, data ?? []];
	}

	async function fetchSets(workspaceId: string): Promise<[string, SetResponse[]]> {
		const { data, error } = await client.GET('/sets', {
			params: { query: { workspace_id: workspaceId } }
		});
		if (error) throw new Error(problemMessage(error, m.calendar_failed_load()));
		return [workspaceId, data ?? []];
	}

	function postToCalendarItem(post: Post): CalendarItem | null {
		if (!post.scheduled_at) return null;
		const accounts = accountsForDestinations(post.workspace_id, post.destinations ?? []);
		const setName = matchingSetName(
			post.workspace_id,
			accounts.map((account) => account.id)
		);
		const initial = initialPostText(post.content);
		return {
			id: post.id,
			key: `post:${post.id}`,
			kind: 'post',
			title: initial || m.calendar_untitled_post(),
			preview: initial || post.content,
			status: post.status,
			scheduledAt: post.scheduled_at,
			randomDelayMinutes: post.random_delay_minutes ?? 0,
			workspaceId: post.workspace_id,
			workspaceName: workspaceName(post.workspace_id),
			accounts,
			platforms: unique(accounts.map((account) => account.platform)),
			setName,
			mediaCount: post.media_ids?.length ?? 0,
			profile: ''
		};
	}

	function publicationToCalendarItem(publication: Publication): CalendarItem | null {
		if (!publication.scheduled_at) return null;
		const renditions = publication.renditions ?? [];
		const accounts = accountsForRenditions(publication.workspace_id, renditions);
		const setName = matchingSetName(
			publication.workspace_id,
			accounts.map((account) => account.id)
		);
		const title =
			publication.title || firstLine(publication.source_text) || m.calendar_untitled_set();
		return {
			id: publication.id,
			key: `publication:${publication.id}`,
			kind: 'publication',
			title,
			preview: firstLine(publication.source_text) || title,
			status: publication.status,
			scheduledAt: publication.scheduled_at,
			randomDelayMinutes: 0,
			workspaceId: publication.workspace_id,
			workspaceName: workspaceName(publication.workspace_id),
			accounts,
			platforms: unique(accounts.map((account) => account.platform)),
			setName,
			mediaCount: publication.media?.length ?? 0,
			profile: publication.content_profile,
			publication
		};
	}

	function accountsForDestinations(workspaceId: string, destinations: PostDestination[]) {
		const byId = accountMap(workspaceId);
		return uniqueById(
			destinations.map((destination) => {
				const account = byId.get(destination.social_account_id);
				if (account) return accountBadge(account);
				return {
					id: destination.social_account_id,
					platform: destination.platform,
					label: platformLabel(destination.platform)
				};
			})
		);
	}

	function accountsForRenditions(workspaceId: string, renditions: Rendition[]) {
		const byId = accountMap(workspaceId);
		return uniqueById(
			renditions.map((rendition) => {
				const account = byId.get(rendition.social_account_id);
				if (account) return accountBadge(account);
				return {
					id: rendition.social_account_id,
					platform: rendition.platform,
					label: platformLabel(rendition.platform)
				};
			})
		);
	}

	function accountMap(workspaceId: string) {
		return new Map(
			(accountsByWorkspace[workspaceId] ?? []).map((account) => [account.id, account])
		);
	}

	function accountBadge(account: SocialAccount): AccountBadge {
		const username = account.account_username?.trim();
		return {
			id: account.id,
			platform: account.platform,
			label: username
				? `@${username.replace(/^@/, '')}`
				: account.slug || platformLabel(account.platform)
		};
	}

	function matchingSetName(workspaceId: string, accountIds: string[]) {
		const key = sortedKey(accountIds);
		if (!key) return '';
		const set = (setsByWorkspace[workspaceId] ?? []).find(
			(candidate) =>
				sortedKey((candidate.accounts ?? []).map((account) => account.social_account_id)) === key
		);
		return set?.name ?? '';
	}

	function toggleWorkspace(workspaceId: string) {
		if (selectedWorkspaceIds.length === 0) {
			selectedWorkspaceIds = workspaces
				.map((workspace) => workspace.id)
				.filter((candidate) => candidate !== workspaceId);
		} else if (selectedWorkspaceIds.includes(workspaceId)) {
			selectedWorkspaceIds = selectedWorkspaceIds.filter((candidate) => candidate !== workspaceId);
		} else {
			selectedWorkspaceIds = [...selectedWorkspaceIds, workspaceId];
		}
		if (selectedWorkspaceIds.length === 0 || selectedWorkspaceIds.length === workspaces.length) {
			selectedWorkspaceIds = [];
		}
	}

	function workspaceSelected(workspaceId: string) {
		return selectedWorkspaceIds.length === 0 || selectedWorkspaceIds.includes(workspaceId);
	}

	function changeMonth(delta: number) {
		currentMonth = startOfMonth(addMonths(currentMonth, delta));
	}

	function goToToday() {
		currentMonth = startOfMonth(new Date());
	}

	function openItem(item: CalendarItem) {
		if (item.kind === 'post') {
			goto(resolve(`/posts/${item.id}`));
		}
	}

	function onDragStart(event: DragEvent, item: CalendarItem) {
		draggingKey = item.key;
		successMessage = '';
		errorMessage = '';
		event.dataTransfer?.setData('text/plain', item.key);
		if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move';
	}

	function onDragEnd() {
		draggingKey = '';
		dropTargetKey = '';
	}

	function onDragOver(event: DragEvent, day: CalendarDay) {
		if (!draggingKey || reschedulingKey) return;
		event.preventDefault();
		dropTargetKey = day.key;
		if (event.dataTransfer) event.dataTransfer.dropEffect = 'move';
	}

	function onDragLeave(day: CalendarDay) {
		if (dropTargetKey === day.key) dropTargetKey = '';
	}

	async function onDrop(event: DragEvent, day: CalendarDay) {
		event.preventDefault();
		const key = event.dataTransfer?.getData('text/plain') || draggingKey;
		const item = allItems.find((candidate) => candidate.key === key);
		draggingKey = '';
		dropTargetKey = '';
		if (!item || dateKey(new Date(item.scheduledAt)) === day.key) return;
		await rescheduleItem(item, day.date);
	}

	async function rescheduleItem(item: CalendarItem, targetDate: Date) {
		const nextScheduledAt = moveDateKeepTime(item.scheduledAt, targetDate);
		const previousPosts = posts;
		const previousPublications = publications;
		reschedulingKey = item.key;
		errorMessage = '';
		successMessage = '';

		if (item.kind === 'post') {
			posts = posts.map((post) =>
				post.id === item.id ? { ...post, scheduled_at: nextScheduledAt } : post
			);
		} else {
			publications = publications.map((publication) =>
				publication.id === item.id ? { ...publication, scheduled_at: nextScheduledAt } : publication
			);
		}

		try {
			if (item.kind === 'post') {
				const { error } = await client.PATCH('/posts/{id}', {
					params: { path: { id: item.id } },
					body: {
						scheduled_at: nextScheduledAt,
						random_delay_minutes: item.randomDelayMinutes
					}
				});
				if (error) throw new Error(problemMessage(error, m.calendar_reschedule_failed()));
			} else if (item.publication) {
				const publication = item.publication;
				const { error } = await client.PUT('/publications/{id}', {
					params: { path: { id: item.id } },
					body: {
						title: publication.title,
						content_profile: publication.content_profile,
						source_text: publication.source_text,
						source_url: publication.source_url ?? '',
						goal: publication.goal ?? '',
						audience: publication.audience ?? '',
						metadata: publication.metadata ?? {},
						scheduled_at: nextScheduledAt
					}
				});
				if (error) throw new Error(problemMessage(error, m.calendar_reschedule_failed()));
			}
			successMessage = m.calendar_rescheduled({
				title: item.title,
				date: formatLongDateTime(nextScheduledAt)
			});
			ui.triggerRefresh();
		} catch (error) {
			posts = previousPosts;
			publications = previousPublications;
			errorMessage = problemMessage(error, m.calendar_reschedule_failed());
		} finally {
			reschedulingKey = '';
		}
	}

	function startOfMonth(date: Date) {
		return new Date(date.getFullYear(), date.getMonth(), 1);
	}

	function addMonths(date: Date, count: number) {
		return new Date(date.getFullYear(), date.getMonth() + count, 1);
	}

	function buildCalendarDays(month: Date, weekStart: number, todayDate: Date) {
		const first = startOfWeek(startOfMonth(month), weekStart);
		const todayKey = dateKey(todayDate);
		const monthValue = month.getMonth();
		return Array.from({ length: 42 }, (_, index): CalendarDay => {
			const date = new Date(first);
			date.setDate(first.getDate() + index);
			return {
				date,
				key: dateKey(date),
				outsideMonth: date.getMonth() !== monthValue,
				today: dateKey(date) === todayKey
			};
		});
	}

	function startOfWeek(date: Date, weekStart: number) {
		const out = new Date(date);
		out.setHours(0, 0, 0, 0);
		const diff = (out.getDay() - weekStart + 7) % 7;
		out.setDate(out.getDate() - diff);
		return out;
	}

	function dateKey(date: Date) {
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		return `${year}-${month}-${day}`;
	}

	function monthKey(date: Date) {
		return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
	}

	function moveDateKeepTime(sourceISO: string, targetDate: Date) {
		const source = new Date(sourceISO);
		const next = new Date(targetDate);
		next.setHours(
			source.getHours(),
			source.getMinutes(),
			source.getSeconds(),
			source.getMilliseconds()
		);
		return next.toISOString();
	}

	function firstLine(text: string) {
		return text.trim().split(/\n+/)[0]?.trim() ?? '';
	}

	function initialPostText(content: string) {
		if (!content.startsWith('__openpost_thread__:')) return firstLine(content);
		try {
			const data = JSON.parse(content.slice('__openpost_thread__:'.length));
			const entries = Array.isArray(data) ? data : Array.isArray(data?.p) ? data.p : [];
			return firstLine(entries[0]?.c ?? '');
		} catch {
			return firstLine(content);
		}
	}

	function unique(values: string[]) {
		return Array.from(new Set(values.filter(Boolean)));
	}

	function uniqueById(accounts: AccountBadge[]) {
		const seen = new Set<string>();
		return accounts.filter((account) => {
			if (!account.id || seen.has(account.id)) return false;
			seen.add(account.id);
			return true;
		});
	}

	function sortedKey(values: string[]) {
		return unique(values).sort().join('|');
	}

	function workspaceName(workspaceId: string) {
		return (
			workspaces.find((workspace) => workspace.id === workspaceId)?.name ??
			m.calendar_unknown_workspace()
		);
	}

	function workspaceInitials(workspaceId: string) {
		const source = workspaceName(workspaceId);
		const parts = source.split(/[\s._-]+/).filter(Boolean);
		return ((parts[0]?.[0] ?? 'W') + (parts[1]?.[0] ?? '')).toUpperCase();
	}

	function workspaceHue(workspaceId: string) {
		let hash = 0;
		for (let index = 0; index < workspaceId.length; index++) {
			hash = workspaceId.charCodeAt(index) + ((hash << 5) - hash);
		}
		return Math.abs(hash) % 360;
	}

	function workspaceDotStyle(workspaceId: string) {
		const hue = workspaceHue(workspaceId);
		return `background-color: oklch(0.62 0.16 ${hue});`;
	}

	function platformLabel(platform: string) {
		const labels: Record<string, string> = {
			x: 'X',
			twitter: 'X',
			mastodon: 'Mastodon',
			bluesky: 'Bluesky',
			linkedin: 'LinkedIn',
			threads: 'Threads',
			facebook: 'Facebook',
			instagram: 'Instagram',
			tiktok: 'TikTok',
			youtube: 'YouTube'
		};
		return labels[platform] ?? platform;
	}

	function formatMonth(date: Date) {
		return date.toLocaleDateString(getLocaleTag(), { month: 'long', year: 'numeric' });
	}

	function formatTime(value: string) {
		return new Date(value).toLocaleTimeString(getLocaleTag(), {
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatLongDateTime(value: string) {
		return new Date(value).toLocaleString(getLocaleTag(), {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function problemMessage(error: unknown, fallback: string) {
		const candidate =
			(error as { detail?: unknown; message?: unknown })?.detail ??
			(error as { message?: unknown })?.message;
		return typeof candidate === 'string' && candidate.trim() ? candidate : fallback;
	}

	function itemTone(item: CalendarItem) {
		if (item.kind === 'publication') {
			return 'border-l-violet-500 bg-violet-50/70 text-violet-950 hover:bg-violet-100 dark:bg-violet-950/30 dark:text-violet-100 dark:hover:bg-violet-950/45';
		}
		return 'border-l-sky-500 bg-sky-50/80 text-sky-950 hover:bg-sky-100 dark:bg-sky-950/30 dark:text-sky-100 dark:hover:bg-sky-950/45';
	}

	function activeWorkspaceButtonClass(workspace: Workspace) {
		return workspaceSelected(workspace.id)
			? 'border-foreground/20 bg-foreground text-background hover:bg-foreground/90'
			: 'border-border bg-background text-muted-foreground hover:bg-muted hover:text-foreground';
	}
</script>

<svelte:head>
	<title>{m.calendar_page_title()}</title>
</svelte:head>

<div class="flex min-h-0 flex-1 flex-col overflow-hidden bg-background">
	<header class="border-b bg-background/95">
		<div class="flex flex-col gap-4 px-4 py-4 lg:px-6">
			<div class="flex flex-col gap-3 xl:flex-row xl:items-start xl:justify-between">
				<div class="min-w-0">
					<div class="flex items-center gap-2 text-sm font-medium text-muted-foreground">
						<CalendarDaysIcon class="size-4" />
						<span>{m.calendar_kicker()}</span>
					</div>
					<div class="mt-1 flex flex-wrap items-end gap-x-3 gap-y-1">
						<h1 class="text-2xl font-semibold tracking-normal text-foreground sm:text-3xl">
							{formatMonth(currentMonth)}
						</h1>
						<div class="flex items-center gap-2 pb-1 text-sm text-muted-foreground">
							<span>{m.calendar_scheduled_summary({ count: monthItems.length })}</span>
							<span class="text-muted-foreground/40">/</span>
							<span
								>{m.calendar_post_set_summary({ posts: postCount, sets: publicationCount })}</span
							>
						</div>
					</div>
				</div>

				<div class="flex flex-wrap items-center gap-2">
					<div class="inline-flex rounded-md border bg-card p-1">
						<Tooltip.Root>
							<Tooltip.Trigger>
								{#snippet child({ props })}
									<Button
										{...props}
										variant="ghost"
										size="icon-sm"
										aria-label={m.calendar_previous_month()}
										onclick={() => changeMonth(-1)}
									>
										<ChevronLeftIcon class="size-4" />
									</Button>
								{/snippet}
							</Tooltip.Trigger>
							<Tooltip.Content>{m.calendar_previous_month()}</Tooltip.Content>
						</Tooltip.Root>
						<Button variant="ghost" size="sm" onclick={goToToday}>{m.calendar_today()}</Button>
						<Tooltip.Root>
							<Tooltip.Trigger>
								{#snippet child({ props })}
									<Button
										{...props}
										variant="ghost"
										size="icon-sm"
										aria-label={m.calendar_next_month()}
										onclick={() => changeMonth(1)}
									>
										<ChevronRightIcon class="size-4" />
									</Button>
								{/snippet}
							</Tooltip.Trigger>
							<Tooltip.Content>{m.calendar_next_month()}</Tooltip.Content>
						</Tooltip.Root>
					</div>

					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							{#snippet child({ props })}
								<Button {...props} variant="outline" class="max-w-64 justify-start">
									<span class="truncate">{selectedWorkspaceLabel}</span>
								</Button>
							{/snippet}
						</DropdownMenu.Trigger>
						<DropdownMenu.Content class="w-64" align="end">
							<DropdownMenu.Label>{m.calendar_workspace_filter()}</DropdownMenu.Label>
							<DropdownMenu.Item onclick={() => (selectedWorkspaceIds = [])} class="gap-2">
								<CheckIcon
									class={cn(
										'size-4',
										selectedWorkspaceIds.length === 0 ? 'opacity-100' : 'opacity-0'
									)}
								/>
								<span>{m.calendar_all_workspaces()}</span>
							</DropdownMenu.Item>
							<DropdownMenu.Separator />
							{#each workspaces as workspace (workspace.id)}
								<DropdownMenu.Item onclick={() => toggleWorkspace(workspace.id)} class="gap-2">
									<CheckIcon
										class={cn(
											'size-4',
											workspaceSelected(workspace.id) ? 'opacity-100' : 'opacity-0'
										)}
									/>
									<span class="h-2 w-2 rounded-full" style={workspaceDotStyle(workspace.id)}></span>
									<span class="truncate">{workspace.name}</span>
								</DropdownMenu.Item>
							{/each}
						</DropdownMenu.Content>
					</DropdownMenu.Root>

					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							{#snippet child({ props })}
								<Button {...props} variant="outline" class="max-w-48 justify-start">
									<span class="truncate">
										{selectedPlatform === 'all'
											? m.calendar_all_platforms()
											: platformLabel(selectedPlatform)}
									</span>
								</Button>
							{/snippet}
						</DropdownMenu.Trigger>
						<DropdownMenu.Content class="w-52" align="end">
							<DropdownMenu.Label>{m.calendar_platform_filter()}</DropdownMenu.Label>
							<DropdownMenu.Item onclick={() => (selectedPlatform = 'all')} class="gap-2">
								<CheckIcon
									class={cn('size-4', selectedPlatform === 'all' ? 'opacity-100' : 'opacity-0')}
								/>
								<span>{m.calendar_all_platforms()}</span>
							</DropdownMenu.Item>
							{#each availablePlatforms as platform (platform)}
								<DropdownMenu.Item onclick={() => (selectedPlatform = platform)} class="gap-2">
									<CheckIcon
										class={cn(
											'size-4',
											selectedPlatform === platform ? 'opacity-100' : 'opacity-0'
										)}
									/>
									<PlatformIcon {platform} class="size-4" />
									<span>{platformLabel(platform)}</span>
								</DropdownMenu.Item>
							{/each}
						</DropdownMenu.Content>
					</DropdownMenu.Root>

					<Tooltip.Root>
						<Tooltip.Trigger>
							{#snippet child({ props })}
								<Button
									{...props}
									variant="outline"
									size="icon"
									aria-label={m.calendar_refresh()}
									disabled={loading}
									onclick={() => loadCalendarData(loadKey)}
								>
									<RefreshCwIcon class={cn('size-4', loading && 'animate-spin')} />
								</Button>
							{/snippet}
						</Tooltip.Trigger>
						<Tooltip.Content>{m.calendar_refresh()}</Tooltip.Content>
					</Tooltip.Root>
				</div>
			</div>

			<div class="flex gap-2 overflow-x-auto pb-1">
				<Button
					variant="outline"
					size="sm"
					class={cn(
						'shrink-0',
						selectedWorkspaceIds.length === 0
							? 'border-foreground/20 bg-foreground text-background hover:bg-foreground/90'
							: ''
					)}
					onclick={() => (selectedWorkspaceIds = [])}
				>
					{m.calendar_all_workspaces()}
				</Button>
				{#each workspaces as workspace (workspace.id)}
					<button
						type="button"
						class={cn(
							'inline-flex h-8 shrink-0 items-center gap-2 rounded-md border px-2.5 text-sm font-medium transition-colors',
							activeWorkspaceButtonClass(workspace)
						)}
						onclick={() => toggleWorkspace(workspace.id)}
					>
						<span class="h-2 w-2 rounded-full" style={workspaceDotStyle(workspace.id)}></span>
						<span class="max-w-36 truncate">{workspace.name}</span>
					</button>
				{/each}
			</div>
		</div>
	</header>

	{#if errorMessage}
		<div
			class="border-b border-destructive/20 bg-destructive/10 px-4 py-2 text-sm text-destructive lg:px-6"
		>
			{errorMessage}
		</div>
	{:else if successMessage}
		<div
			class="border-b border-emerald-500/20 bg-emerald-500/10 px-4 py-2 text-sm text-emerald-800 lg:px-6 dark:text-emerald-200"
		>
			{successMessage}
		</div>
	{/if}

	<main class="min-h-0 flex-1 overflow-auto px-3 py-3 lg:px-6 lg:py-5">
		<section
			class="month-shell grid h-full min-h-[720px] min-w-[980px] grid-rows-[auto_minmax(0,1fr)] overflow-hidden rounded-lg border bg-card shadow-sm"
			aria-label={m.calendar_month_grid()}
		>
			<div class="grid grid-cols-7 border-b bg-muted/45">
				{#each weekdayLabels as label (label)}
					<div class="px-3 py-2 text-xs font-medium tracking-normal text-muted-foreground">
						{label}
					</div>
				{/each}
			</div>
			<div class="grid min-h-0 grid-cols-7 grid-rows-6">
				{#each days as day (day.key)}
					{@const dayItems = itemsByDay.get(day.key) ?? []}
					<section
						role="gridcell"
						tabindex="-1"
						class={cn(
							'group/day relative flex min-h-0 flex-col border-r border-b bg-background/70 p-2 transition-colors last:border-r-0',
							day.outsideMonth && 'bg-muted/25 text-muted-foreground',
							day.today && 'bg-primary/[0.035]',
							dropTargetKey === day.key && 'bg-primary/10 ring-2 ring-primary ring-inset'
						)}
						ondragover={(event) => onDragOver(event, day)}
						ondragleave={() => onDragLeave(day)}
						ondrop={(event) => onDrop(event, day)}
					>
						<div class="mb-2 flex items-start justify-between gap-2">
							<div class="flex items-center gap-2">
								<span
									class={cn(
										'flex size-7 items-center justify-center rounded-md text-sm font-medium',
										day.today && 'bg-primary text-primary-foreground',
										day.outsideMonth && !day.today && 'text-muted-foreground'
									)}
								>
									{day.date.getDate()}
								</span>
								{#if dayItems.length > 0}
									<Badge class="bg-muted text-muted-foreground">
										{m.calendar_day_item_count({ count: dayItems.length })}
									</Badge>
								{/if}
							</div>
							{#if loading}
								<Loader2Icon class="mt-1 size-3.5 animate-spin text-muted-foreground" />
							{/if}
						</div>

						<div class="calendar-day-scroll min-h-0 flex-1 space-y-1.5 overflow-y-auto pr-1">
							{#if loading}
								{#each Array(3) as _, index (index)}
									<div class="h-16 rounded-md border bg-muted/45"></div>
								{/each}
							{:else if dayItems.length === 0}
								<div
									class="flex h-full min-h-24 items-center justify-center rounded-md border border-dashed border-transparent text-xs text-muted-foreground/55 group-hover/day:border-border"
								>
									{m.calendar_empty_day()}
								</div>
							{:else}
								{#each dayItems as item (item.key)}
									<button
										type="button"
										draggable={true}
										class={cn(
											'w-full rounded-md border border-l-4 border-border/70 p-2 text-left shadow-xs transition-all hover:-translate-y-0.5 hover:shadow-sm focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none',
											itemTone(item),
											draggingKey === item.key && 'opacity-50',
											reschedulingKey === item.key && 'pointer-events-none opacity-60'
										)}
										aria-label={item.kind === 'post'
											? m.calendar_open_post({ title: item.title })
											: m.calendar_set_card({ title: item.title })}
										ondragstart={(event) => onDragStart(event, item)}
										ondragend={onDragEnd}
										onclick={() => openItem(item)}
									>
										<div class="flex items-start gap-2">
											<div
												class="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-md text-[10px] font-semibold text-white shadow-xs"
												style={workspaceDotStyle(item.workspaceId)}
											>
												{workspaceInitials(item.workspaceId)}
											</div>
											<div class="min-w-0 flex-1">
												<div
													class="mb-1 flex items-center gap-1.5 text-[11px] font-medium text-current/70"
												>
													<ClockIcon class="size-3" />
													<span>{formatTime(item.scheduledAt)}</span>
													<span class="truncate">{item.workspaceName}</span>
												</div>
												<div class="line-clamp-2 text-sm leading-snug font-semibold">
													{item.title}
												</div>
												{#if item.preview && item.preview !== item.title}
													<div class="mt-0.5 line-clamp-2 text-xs leading-snug text-current/70">
														{item.preview}
													</div>
												{/if}
											</div>
										</div>
										<div class="mt-2 flex flex-wrap items-center gap-1">
											<span
												class="inline-flex items-center gap-1 rounded-sm bg-background/65 px-1.5 py-0.5 text-[11px] font-medium text-current/75 ring-1 ring-current/10"
											>
												{#if item.kind === 'publication'}
													<LayersIcon class="size-3" />
													{m.calendar_set_label()}
												{:else}
													{m.calendar_post_label()}
												{/if}
											</span>
											{#if item.setName}
												<span
													class="inline-flex max-w-36 items-center gap-1 rounded-sm bg-background/65 px-1.5 py-0.5 text-[11px] font-medium text-current/75 ring-1 ring-current/10"
												>
													<LayersIcon class="size-3" />
													<span class="truncate">{item.setName}</span>
												</span>
											{/if}
											{#each item.accounts.slice(0, 3) as account (account.id)}
												<span
													class="inline-flex max-w-32 items-center gap-1 rounded-sm bg-background/65 px-1.5 py-0.5 text-[11px] font-medium text-current/75 ring-1 ring-current/10"
													title={`${platformLabel(account.platform)} ${account.label}`}
												>
													<PlatformIcon platform={account.platform} class="size-3" />
													<span class="truncate">{account.label}</span>
												</span>
											{/each}
											{#if item.accounts.length > 3}
												<span
													class="inline-flex rounded-sm bg-background/65 px-1.5 py-0.5 text-[11px] font-medium text-current/75 ring-1 ring-current/10"
												>
													{m.calendar_more_accounts({ count: item.accounts.length - 3 })}
												</span>
											{/if}
											{#if reschedulingKey === item.key}
												<Loader2Icon class="ml-auto size-3.5 animate-spin text-current/60" />
											{/if}
										</div>
									</button>
								{/each}
							{/if}
						</div>
					</section>
				{/each}
			</div>
		</section>

		{#if !loading && monthItems.length === 0}
			<div class="mx-auto mt-6 max-w-md rounded-lg border bg-card p-6 text-center shadow-sm">
				<CalendarDaysIcon class="mx-auto size-8 text-muted-foreground" />
				<h2 class="mt-3 text-base font-semibold">{m.calendar_no_scheduled_title()}</h2>
				<p class="mt-1 text-sm text-muted-foreground">{m.calendar_no_scheduled_body()}</p>
				<Button class="mt-4" onclick={() => goto(resolve('/'))}>{m.calendar_create_post()}</Button>
			</div>
		{/if}
	</main>
</div>

<style>
	.month-shell {
		grid-template-rows: auto minmax(0, 1fr);
	}

	.calendar-day-scroll {
		scrollbar-width: thin;
		scrollbar-color: color-mix(in oklch, var(--muted-foreground) 30%, transparent) transparent;
	}
</style>
