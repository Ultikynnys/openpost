<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { isNavigationItemActive, mobileNavigation } from '$lib/app-navigation';
	import { ui } from '$lib/stores/ui.svelte';
	import { m } from '$lib/paraglide/messages';
	import CalendarIcon from 'lucide-svelte/icons/calendar-days';
	import ComposeIcon from 'lucide-svelte/icons/plus';
	import PostsIcon from 'lucide-svelte/icons/files';
	import MediaIcon from 'lucide-svelte/icons/images';
	import AccountsIcon from 'lucide-svelte/icons/users';

	const items = mobileNavigation;
	const pathname = $derived(page.url.pathname);

	function iconFor(id: (typeof items)[number]['id']) {
		switch (id) {
			case 'calendar':
				return CalendarIcon;
			case 'posts':
				return PostsIcon;
			case 'media':
				return MediaIcon;
			case 'accounts':
				return AccountsIcon;
			default:
				return ComposeIcon;
		}
	}

	function labelFor(id: (typeof items)[number]['id']) {
		switch (id) {
			case 'new':
				return m.sidebar_new();
			case 'calendar':
				return m.sidebar_calendar();
			case 'posts':
				return m.sidebar_activity();
			case 'media':
				return m.sidebar_media();
			case 'accounts':
				return m.sidebar_accounts();
			default:
				return '';
		}
	}
</script>

<nav
	class="fixed inset-x-0 bottom-0 z-30 border-t bg-background/96 px-2 pt-1 pb-[max(0.35rem,env(safe-area-inset-bottom))] backdrop-blur-md md:hidden"
	aria-label={m.sidebar_primary_navigation()}
>
	<ul class="grid grid-cols-5">
		{#each items as item (item.id)}
			{@const Icon = iconFor(item.id)}
			{@const active = isNavigationItemActive(item, pathname)}
			<li>
				<button
					type="button"
					class={[
						'flex min-h-14 w-full flex-col items-center justify-center gap-1 rounded-md px-1 text-[0.68rem] font-medium transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none',
						item.id === 'new'
							? 'text-primary'
							: active
								? 'text-foreground'
								: 'text-muted-foreground'
					]}
					onclick={() => {
						if (item.id === 'new') ui.startNewPost();
						goto(resolve(item.href as '/'));
					}}
					aria-current={active ? 'page' : undefined}
				>
					<span
						class={item.id === 'new'
							? 'flex size-8 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-sm'
							: 'flex size-5 items-center justify-center'}
					>
						<Icon class={item.id === 'new' ? 'size-5' : 'size-4'} />
					</span>
					<span>{labelFor(item.id)}</span>
				</button>
			</li>
		{/each}
	</ul>
</nav>
