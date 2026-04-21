<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api, type Transaction, type Category, type Account, type MerchantIcon as MerchantIconData } from '$lib/api';
	import MerchantIcon from '$lib/MerchantIcon.svelte';
	import type { SvelteComponent, ComponentType } from 'svelte';

	let IconPicker: ComponentType<SvelteComponent> | null = null;
	async function ensureIconPicker() {
		if (!IconPicker) {
			IconPicker = (await import('$lib/IconPicker.svelte')).default;
		}
	}

	let transactions: Transaction[] = [];
	let categories: Category[] = [];
	let accounts: Account[] = [];
	let merchantIconOverrides: MerchantIconData[] = [];
	let loading = true;
	let error = '';

	// Filters (bound to URL search params)
	let from = '';
	let to = '';
	let category = '';
	let account = '';
	let q = '';
	let offset = 0;
	const limit = 50;

	onMount(async () => {
		// Restore filters from URL
		const p = $page.url.searchParams;
		from = p.get('from') ?? '';
		to = p.get('to') ?? '';
		category = p.get('category') ?? '';
		account = p.get('account') ?? '';
		q = p.get('q') ?? '';
		offset = Number(p.get('offset') ?? 0);

		try {
			[categories, accounts, merchantIconOverrides] = await Promise.all([
				api.categories(),
				api.accounts(),
				api.merchantIcons()
			]);
			await loadTransactions();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load';
			loading = false;
		}
	});

	async function loadTransactions() {
		loading = true;
		error = '';
		try {
			transactions = await api.transactions({ from, to, category, account, q, limit, offset });
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load transactions';
		} finally {
			loading = false;
		}
	}

	function applyFilters() {
		offset = 0;
		const params = new URLSearchParams();
		if (from) params.set('from', from);
		if (to) params.set('to', to);
		if (category) params.set('category', category);
		if (account) params.set('account', account);
		if (q) params.set('q', q);
		goto(`?${params.toString()}`, { replaceState: true });
		loadTransactions();
	}

	function nextPage() {
		offset += limit;
		loadTransactions();
	}

	function prevPage() {
		offset = Math.max(0, offset - limit);
		loadTransactions();
	}

	$: merchantIconMap = Object.fromEntries(
		merchantIconOverrides.map((m) => [m.merchant_name.toLowerCase(), m.icon_slug])
	);

	function merchantSlug(txn: Transaction): string | null {
		const name = (txn.merchant_clean ?? txn.merchant).toLowerCase();
		return merchantIconMap[name] ?? null;
	}

	// Merchant icon picker state
	let merchantPickerOpen = false;
	let merchantPickerName = '';
	let merchantPickerSlug: string | null = null;

	async function openMerchantPicker(txn: Transaction) {
		merchantPickerName = (txn.merchant_clean ?? txn.merchant).trim().toLowerCase();
		merchantPickerSlug = merchantSlug(txn);
		await ensureIconPicker();
		merchantPickerOpen = true;
	}

	async function handleMerchantIconSelect(e: CustomEvent<string | null>) {
		const slug = e.detail;
		const name = merchantPickerName;
		try {
			if (slug) {
				await api.setMerchantIcon(name, slug);
				const existing = merchantIconOverrides.findIndex((m) => m.merchant_name === name);
				if (existing >= 0) {
					merchantIconOverrides = merchantIconOverrides.map((m) =>
						m.merchant_name === name ? { ...m, icon_slug: slug } : m
					);
				} else {
					merchantIconOverrides = [...merchantIconOverrides, { merchant_name: name, icon_slug: slug, updated_at: '' }];
				}
			} else {
				await api.deleteMerchantIcon(name);
				merchantIconOverrides = merchantIconOverrides.filter((m) => m.merchant_name !== name);
			}
		} catch { /* ignore */ }
		merchantPickerName = '';
	}

	function formatAmount(amount: number) {
		const abs = Math.abs(amount);
		return (amount < 0 ? '-' : '+') + '$' + abs.toFixed(2);
	}
</script>

<svelte:head>
	<title>Transactions — miser</title>
</svelte:head>

<div class="page">
	<h1>Transactions</h1>

	<form class="filters" on:submit|preventDefault={applyFilters}>
		<input type="date" bind:value={from} placeholder="From" title="From date" />
		<input type="date" bind:value={to} placeholder="To" title="To date" />
		<select bind:value={category}>
			<option value="">All categories</option>
			{#each categories as cat}
				<option value={cat.name}>{cat.name}</option>
			{/each}
		</select>
		<select bind:value={account}>
			<option value="">All accounts</option>
			{#each accounts as acct}
				<option value={acct.name}>{acct.name}</option>
			{/each}
		</select>
		<input type="text" bind:value={q} placeholder="Search merchant..." />
		<button type="submit">Filter</button>
	</form>

	{#if loading}
		<p class="state">Loading...</p>
	{:else if error}
		<p class="state error">{error}</p>
	{:else if transactions.length === 0}
		<p class="state">No transactions found.</p>
	{:else}
		<table>
			<thead>
				<tr>
					<th>Date</th>
					<th>Merchant</th>
					<th>Category</th>
					<th>Account</th>
					<th class="right">Amount</th>
				</tr>
			</thead>
			<tbody>
				{#each transactions as txn}
					<tr>
						<td class="muted">{txn.date}</td>
						<td>
								<div class="merchant-cell">
									<button class="icon-btn" on:click={() => openMerchantPicker(txn)}>
										<MerchantIcon merchant={txn.merchant_clean ?? txn.merchant} size={28} iconSlug={merchantIconMap[(txn.merchant_clean ?? txn.merchant).toLowerCase()] ?? null} />
									</button>
									<span>{txn.merchant_clean ?? txn.merchant}</span>
								</div>
							</td>
						<td class="muted">{txn.category_name || 'Uncategorized'}</td>
						<td class="muted">{txn.account_name}</td>
						<td class="right mono" class:income={txn.amount > 0} class:expense={txn.amount < 0}>
							{formatAmount(txn.amount)}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>

		<div class="pagination">
			<button on:click={prevPage} disabled={offset === 0}>← Prev</button>
			<span class="muted">Showing {offset + 1}–{offset + transactions.length}</span>
			<button on:click={nextPage} disabled={transactions.length < limit}>Next →</button>
		</div>
	{/if}
</div>

{#if IconPicker}
	<svelte:component this={IconPicker} bind:open={merchantPickerOpen} current={merchantPickerSlug} on:select={handleMerchantIconSelect} />
{/if}

<style>
	.page {
		max-width: 1600px;
	}

	h1 {
		font-family: var(--font-display);
		font-size: 36px;
		font-weight: 600;
		color: var(--color-text);
		margin-bottom: 28px;
		letter-spacing: -0.3px;
	}

	/* ── Filters ──────────────────────────────────────── */
	.filters {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-bottom: 28px;
		align-items: center;
	}

	.filters input[type='text'] {
		flex: 2;
		min-width: 160px;
	}

	.filters input,
	.filters select {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		color: var(--color-text);
		padding: 8px 10px;
		font-family: var(--font-sans);
		font-size: 13px;
		font-weight: 400;
		outline: none;
		transition: border-color 0.12s;
	}

	.filters input::placeholder {
		color: var(--color-text-muted);
	}

	.filters input:focus,
	.filters select:focus {
		border-color: var(--color-accent);
	}

	.filters button {
		background: var(--color-accent);
		border: none;
		border-radius: var(--radius);
		color: #ffffff;
		padding: 8px 18px;
		font-family: var(--font-sans);
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		letter-spacing: 0.02em;
		transition: opacity 0.12s;
	}

	.filters button:hover {
		opacity: 0.88;
	}

	/* ── States ───────────────────────────────────────── */
	.state {
		color: var(--color-text-muted);
		margin-top: 32px;
	}

	.error {
		color: var(--color-expense);
	}

	/* ── Table ────────────────────────────────────────── */
	table {
		width: 100%;
		border-collapse: collapse;
	}

	th {
		text-align: left;
		font-size: 11px;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--color-text-muted);
		padding: 10px 14px;
		border-bottom: 1px solid var(--color-border);
		background: var(--color-bg);
	}

	td {
		padding: 0 14px;
		height: 48px;
		vertical-align: middle;
		border-bottom: 1px solid var(--color-border);
		font-size: 14px;
	}

	tbody tr:nth-child(even) td {
		background: var(--color-surface-alt);
	}

	tbody tr:last-child td {
		border-bottom: none;
	}

	tbody tr:hover td {
		background: var(--color-accent-light);
	}

	.muted {
		color: var(--color-text-muted);
		font-size: 13px;
	}

	.right {
		text-align: right;
	}

	.merchant-cell {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.icon-btn {
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		border-radius: var(--radius);
		flex-shrink: 0;
		transition: opacity 0.12s;
	}

	.icon-btn:hover {
		opacity: 0.7;
	}

	.mono {
		font-family: var(--font-mono);
	}

	.income {
		color: var(--color-income);
	}

	.expense {
		color: var(--color-expense);
	}

	/* ── Pagination ───────────────────────────────────── */
	.pagination {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		margin-top: 24px;
		font-size: 13px;
		color: var(--color-text-muted);
	}

	.pagination button {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius);
		color: var(--color-text);
		padding: 7px 14px;
		font-family: var(--font-sans);
		font-size: 13px;
		cursor: pointer;
		transition:
			border-color 0.12s,
			color 0.12s;
	}

	.pagination button:hover:not(:disabled) {
		border-color: var(--color-accent);
		color: var(--color-accent);
	}

	.pagination button:disabled {
		opacity: 0.35;
		cursor: default;
	}
</style>
