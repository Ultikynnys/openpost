<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import InlineNotice from '$lib/components/inline-notice.svelte';
	import StandaloneShell from '$lib/components/standalone-shell.svelte';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import CheckCircleIcon from 'lucide-svelte/icons/check-circle-2';
	import { m } from '$lib/paraglide/messages';
	import { onboardingPathForPlan } from '$lib/billing';

	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let error = $state('');
	let isLoading = $state(false);
	let registrationSuccess = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';

		if (password !== confirmPassword) {
			error = m.auth_register_password_mismatch();
			return;
		}

		if (password.length < 8) {
			error = m.auth_register_password_short();
			return;
		}

		isLoading = true;

		const result = await auth.register(email, password);

		if (result.success) {
			registrationSuccess = true;
			goto(resolve(onboardingPathForPlan(page.url.searchParams.get('plan')) as '/'));
		} else {
			error = result.error || m.auth_register_failed();
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>{m.auth_register_title()}</title>
</svelte:head>

{#if registrationSuccess}
	{#snippet successIcon()}
		<CheckCircleIcon class="size-6 text-emerald-600 dark:text-emerald-400" />
	{/snippet}
	<StandaloneShell
		title={m.auth_register_success_title()}
		description={m.auth_register_success_description()}
		icon={successIcon}
	/>
{:else}
	<StandaloneShell
		title={m.auth_register_heading()}
		description={m.auth_register_description()}
		logoHref="/"
	>
		{#if error}
			<InlineNotice tone="error" message={error} class="mb-4" />
		{/if}

		<form onsubmit={handleSubmit} class="space-y-4">
			<div class="space-y-2">
				<Label for="email">{m.common_email()}</Label>
				<Input
					type="email"
					id="email"
					bind:value={email}
					required
					placeholder={m.auth_email_placeholder()}
				/>
			</div>

			<div class="space-y-2">
				<Label for="password">{m.common_password()}</Label>
				<Input
					type="password"
					id="password"
					bind:value={password}
					required
					placeholder={m.auth_password_min_placeholder()}
				/>
			</div>

			<div class="space-y-2">
				<Label for="confirmPassword">{m.auth_confirm_password()}</Label>
				<Input
					type="password"
					id="confirmPassword"
					bind:value={confirmPassword}
					required
					placeholder={m.auth_password_confirm_placeholder()}
				/>
			</div>

			<Button type="submit" disabled={isLoading} class="w-full gap-2">
				{#if isLoading}
					<LoaderIcon class="h-4 w-4 animate-spin" />
					{m.auth_register_loading()}
				{:else}
					{m.auth_register_submit()}
				{/if}
			</Button>
		</form>

		<p class="mt-6 text-center text-sm text-muted-foreground">
			{m.auth_register_have_account()}
			<a
				href={resolve('/login')}
				class="inline-flex min-h-11 items-center px-1 font-medium text-primary hover:underline"
				>{m.auth_register_sign_in()}</a
			>
		</p>
	</StandaloneShell>
{/if}
