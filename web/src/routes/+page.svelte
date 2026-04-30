<script lang="ts">
	import { onMount } from 'svelte';
	import { marked } from 'marked';
	import { api, type TrendsResponse, type Transaction, type Report, type Category, type MerchantIcon as MerchantIconData } from '$lib/api';
	import MerchantIcon from '$lib/MerchantIcon.svelte';
	import CategoryIcon from '$lib/CategoryIcon.svelte';
	import type { SvelteComponent, ComponentType } from 'svelte';

	marked.setOptions({ gfm: true, breaks: true });

	let IconPicker: ComponentType<SvelteComponent> | null = null;
	async function ensureIconPicker() {
		if (!IconPicker) {
			IconPicker = (await import('$lib/IconPicker.svelte')).default;
		}
	}

	let trends: TrendsResponse | null = null;
	let recentTxns: Transaction[] = [];
	let report: Report | null = null;
	let categories: Category[] = [];
	let merchantIconOverrides: MerchantIconData[] = [];
	let prevMTDTotal: number | null = null;
	let loading = true;
	let error = '';

	// Category icon picker state
	let catPickerOpen = false;
	let catPickerId = '';
	let catPickerSlug: string | null = null;

	// Merchant icon picker state
	let merchantPickerOpen = false;
	let merchantPickerName = '';
	let merchantPickerSlug: string | null = null;

	function ymd(year: number, monthIdx: number, day: number) {
		return `${year}-${String(monthIdx + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
	}

	function previousPeriodRange(today: Date) {
		const year = today.getFullYear();
		const monthIdx = today.getMonth();
		const day = today.getDate();
		const prevYear = monthIdx === 0 ? year - 1 : year;
		const prevMonthIdx = monthIdx === 0 ? 11 : monthIdx - 1;
		const prevMonthLastDay = new Date(prevYear, prevMonthIdx + 1, 0).getDate();
		const prevDay = Math.min(day, prevMonthLastDay);
		return {
			from: ymd(prevYear, prevMonthIdx, 1),
			to: ymd(prevYear, prevMonthIdx, prevDay)
		};
	}

	function sumExpenses(txns: Transaction[]) {
		return txns
			.filter((t) => t.amount < 0)
			.reduce((sum, t) => sum + Math.abs(t.amount), 0);
	}

	onMount(async () => {
		try {
			const prevRange = previousPeriodRange(new Date());
			let prevTxns: Transaction[];
			[trends, recentTxns, report, categories, merchantIconOverrides, prevTxns] = await Promise.all([
				api.trends(),
				api.transactions({ limit: 10 }),
				api.latestReport(),
				api.categories(),
				api.merchantIcons(),
				api.transactions({ from: prevRange.from, to: prevRange.to, limit: 10000 })
			]);
			prevMTDTotal = sumExpenses(prevTxns);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load dashboard';
		} finally {
			loading = false;
		}
	});

	$: categoryMap = new Map(categories.map((c) => [c.name, c]));

	// Map merchant name → icon slug override (from DB)
	$: merchantIconMap = Object.fromEntries(
		merchantIconOverrides.map((m) => [m.merchant_name.toLowerCase(), m.icon_slug])
	);

	function merchantSlug(txn: Transaction): string | null {
		const name = (txn.merchant_clean ?? txn.merchant).toLowerCase();
		return merchantIconMap[name] ?? null;
	}

	async function openCatPicker(catName: string) {
		const cat = categoryMap.get(catName);
		if (!cat) return;
		catPickerId = cat.id;
		catPickerSlug = cat.icon ?? null;
		await ensureIconPicker();
		catPickerOpen = true;
	}

	async function handleCatIconSelect(e: CustomEvent<string | null>) {
		const slug = e.detail;
		try {
			await api.updateCategoryIcon(catPickerId, slug);
			categories = categories.map((c) =>
				c.id === catPickerId ? { ...c, icon: slug ?? undefined } : c
			);
		} catch { /* ignore */ }
		catPickerId = '';
	}

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

	function amountClass(amount: number) {
		return amount < 0 ? 'expense' : 'income';
	}

	$: topCategories = trends?.current.slice(0, 5) ?? [];
	$: budgetMap = Object.fromEntries(
		(trends?.budgets ?? []).map((b) => [b.category, b.budget])
	);
	$: heroTotal =
		trends?.current
			.filter((cat) => cat.total < 0)
			.reduce((sum, cat) => sum + Math.abs(cat.total), 0) ?? 0;

	$: momDeltaPct =
		prevMTDTotal !== null && prevMTDTotal > 0
			? ((heroTotal - prevMTDTotal) / prevMTDTotal) * 100
			: null;
</script>

<svelte:head>
	<title>Dashboard — miser</title>
</svelte:head>

<div>
	<header class="stats">
		<div class="stat">
			<div class="stat-label">{trends?.current_month?.toUpperCase() ?? 'LOADING'} — SPENT</div>
			<div class="stat-value">
				${heroTotal.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
			</div>
			<div class="stat-sub">
				{#if !loading && momDeltaPct !== null}
					<span class="delta {momDeltaPct >= 0 ? 'expense' : 'income'}">
						{momDeltaPct >= 0 ? '▲' : '▼'} {Math.abs(momDeltaPct).toFixed(1)}%
					</span>
					vs same point last month (${prevMTDTotal!.toLocaleString('en-US', { maximumFractionDigits: 0 })})
				{:else if !loading}
					No prior month data
				{/if}
			</div>
		</div>
	</header>

	<hr class="divider" />

	{#if loading}
		<p class="loading">Loading...</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else}
		<div class="grid">
			<!-- Top spending categories -->
			<section class="card">
				<h2>Top Categories</h2>
				<ul class="category-list">
					{#each topCategories as cat}
						{@const budget = budgetMap[cat.category]}
						{@const pct = budget ? Math.min(100, (Math.abs(cat.total) / budget) * 100) : null}
						{@const catData = categoryMap.get(cat.category)}
						<li>
							<div class="cat-row">
								<div class="cat-name-wrap">
									<CategoryIcon
										name={cat.category}
										iconSlug={catData?.icon ?? null}
										size={24}
										on:click={() => openCatPicker(cat.category)}
									/>
									<span class="cat-name">{cat.category}</span>
								</div>
								<span class="cat-amount {amountClass(cat.total)}">{formatAmount(cat.total)}</span>
							</div>
							{#if pct !== null}
								<div class="budget-bar-row">
									<div class="budget-bar">
										<div
											class="budget-fill"
											class:over={pct >= 100}
											style="width: {pct}%"
										></div>
									</div>
									<span class="budget-pct" class:over={pct >= 100}>{pct.toFixed(0)}%</span>
								</div>
							{/if}
						</li>
					{/each}
				</ul>
			</section>

			<!-- Recent transactions -->
			<section class="card">
				<h2>Recent Transactions</h2>
				<ul class="txn-list">
					{#each recentTxns as txn}
						<li class="txn-row">
							<div class="txn-left">
								<button class="icon-btn" on:click={() => openMerchantPicker(txn)}>
									<MerchantIcon merchant={txn.merchant_clean ?? txn.merchant} size={34} iconSlug={merchantIconMap[(txn.merchant_clean ?? txn.merchant).toLowerCase()] ?? null} />
								</button>
								<div class="txn-info">
									<span class="txn-merchant">{txn.merchant_clean ?? txn.merchant}</span>
									<span class="txn-cat">{txn.category_name || 'Uncategorized'}</span>
								</div>
							</div>
							<div class="txn-right">
								<span class="txn-amount {amountClass(txn.amount)}">{formatAmount(txn.amount)}</span>
								<span class="txn-date">{txn.date}</span>
							</div>
						</li>
					{/each}
				</ul>
				<a href="/transactions" class="view-all">View all transactions →</a>
			</section>
		</div>

		{#if report}
			<article class="narrative">
				<h2>Monthly Report — {report.year}/{String(report.month).padStart(2, '0')}</h2>
				<div class="narrative-body">
					{@html marked.parse(report.narrative)}
				</div>
			</article>
		{/if}
	{/if}
</div>

{#if IconPicker}
	<svelte:component this={IconPicker} bind:open={catPickerOpen} current={catPickerSlug} on:select={handleCatIconSelect} />
	<svelte:component this={IconPicker} bind:open={merchantPickerOpen} current={merchantPickerSlug} on:select={handleMerchantIconSelect} />
{/if}

<style>
	/* ── Stats band (extensible row of stat cells) ────── */
	.stats {
		display: flex;
		flex-wrap: wrap;
		gap: 48px 64px;
		margin-bottom: 32px;
	}

	.stat {
		min-width: 0;
	}

	.stat-label {
		font-size: 0.647rem;
		font-weight: 500;
		letter-spacing: 0.14em;
		color: var(--color-accent);
		margin-bottom: 8px;
	}

	.stat-value {
		font-family: var(--font-display);
		font-size: 4.706rem;
		font-weight: 600;
		color: var(--color-text);
		line-height: 1;
		letter-spacing: -1px;
	}

	.stat-sub {
		font-size: 0.882rem;
		color: var(--color-text-muted);
		margin-top: 10px;
		min-height: 1.5em;
	}

	.delta {
		font-family: var(--font-mono);
		font-weight: 500;
		margin-right: 6px;
	}

	.divider {
		border: none;
		border-top: 1px solid var(--color-border);
		margin-bottom: 32px;
	}

	/* ── Loading / error states ───────────────────────── */
	.loading,
	.error {
		color: var(--color-text-muted);
		margin-top: 32px;
	}

	.error {
		color: var(--color-expense);
	}

	/* ── Grid ─────────────────────────────────────────── */
	.grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 24px;
	}

	@media (min-width: 700px) {
		.grid {
			grid-template-columns: 1fr 1fr;
		}
	}

	.card {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: 24px;
	}

	/* ── Section headers ──────────────────────────────── */
	h2 {
		font-size: 0.647rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		color: var(--color-text-muted);
		margin-bottom: 20px;
	}

	/* ── Categories ───────────────────────────────────── */
	.category-list {
		list-style: none;
		display: flex;
		flex-direction: column;
	}

	.category-list li {
		padding: 12px 0;
		border-bottom: 1px solid var(--color-border);
	}

	.category-list li:last-child {
		border-bottom: none;
	}

	.cat-row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin-bottom: 7px;
	}

	.cat-name-wrap {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.cat-name {
		font-size: 0.824rem;
		font-weight: 500;
		color: var(--color-text);
	}

	.cat-amount {
		font-family: var(--font-mono);
		font-size: 0.824rem;
	}

	.budget-bar-row {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.budget-bar {
		flex: 1;
		height: 6px;
		background: var(--color-surface-alt);
		border-radius: 3px;
		overflow: hidden;
	}

	.budget-fill {
		height: 100%;
		background: var(--color-accent);
		border-radius: 3px;
		transition: width 0.35s ease;
	}

	.budget-fill.over {
		background: var(--color-expense);
	}

	.budget-pct {
		font-family: var(--font-mono);
		font-size: 0.647rem;
		color: var(--color-text-muted);
		min-width: 32px;
		text-align: right;
	}

	.budget-pct.over {
		color: var(--color-expense);
	}

	/* ── Clickable icon wrapper ───────────────────────── */
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

	/* ── Transactions ─────────────────────────────────── */
	.txn-list {
		list-style: none;
		display: flex;
		flex-direction: column;
	}

	.txn-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 11px 8px;
		margin: 0 -8px;
		border-bottom: 1px solid var(--color-border);
		border-radius: var(--radius);
		transition: background 0.1s;
	}

	.txn-row:hover {
		background: var(--color-surface-alt);
	}

	.txn-row:last-child {
		border-bottom: none;
	}

	.txn-left {
		display: flex;
		flex-direction: row;
		align-items: center;
		gap: 12px;
	}

	.txn-info {
		display: flex;
		flex-direction: column;
		gap: 3px;
	}

	.txn-merchant {
		font-size: 0.882rem;
		font-weight: 500;
		color: var(--color-text);
	}

	.txn-cat {
		font-size: 0.765rem;
		color: var(--color-text-muted);
	}

	.txn-right {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 3px;
	}

	.txn-amount {
		font-family: var(--font-mono);
		font-size: 0.824rem;
	}

	.txn-date {
		font-size: 0.706rem;
		color: var(--color-text-muted);
	}

	.view-all {
		display: block;
		margin-top: 16px;
		font-size: 0.765rem;
		color: var(--color-text-muted);
		text-align: right;
		transition: color 0.12s;
	}

	.view-all:hover {
		color: var(--color-accent);
	}

	/* ── Narrative (full-width band below grid) ───────── */
	.narrative {
		margin-top: 32px;
	}

	.narrative-body {
		border-left: 2px solid var(--color-accent);
		padding-left: 24px;
		margin-top: 4px;
		font-family: var(--font-display);
		font-size: 1rem;
		line-height: 1.75;
		color: var(--color-text-muted);
	}

	@media (min-width: 1100px) {
		.narrative-body {
			column-count: 2;
			column-gap: 56px;
		}

		.narrative-body :global(p),
		.narrative-body :global(li) {
			break-inside: avoid;
		}
	}

	@media (min-width: 1800px) {
		.narrative-body {
			column-count: 3;
		}
	}

	.narrative-body :global(p) {
		margin-bottom: 12px;
	}

	.narrative-body :global(p:last-child) {
		margin-bottom: 0;
	}

	.narrative-body :global(h1),
	.narrative-body :global(h2),
	.narrative-body :global(h3),
	.narrative-body :global(h4) {
		font-family: var(--font-display);
		font-weight: 600;
		color: var(--color-text);
		letter-spacing: -0.2px;
		margin: 18px 0 8px;
		text-transform: none;
	}

	.narrative-body :global(h1) { font-size: 1.412rem; }
	.narrative-body :global(h2) { font-size: 1.176rem; }
	.narrative-body :global(h3) { font-size: 1rem; }
	.narrative-body :global(h4) { font-size: 0.882rem; }

	.narrative-body :global(h1:first-child),
	.narrative-body :global(h2:first-child),
	.narrative-body :global(h3:first-child),
	.narrative-body :global(h4:first-child) {
		margin-top: 0;
	}

	.narrative-body :global(ul),
	.narrative-body :global(ol) {
		padding-left: 22px;
		margin-bottom: 12px;
	}

	.narrative-body :global(li) {
		margin-bottom: 4px;
	}

	.narrative-body :global(strong) {
		font-weight: 600;
		color: var(--color-text);
	}

	.narrative-body :global(em) {
		font-style: italic;
	}

	.narrative-body :global(code) {
		font-family: var(--font-mono);
		font-size: 0.9em;
		background: var(--color-surface-alt);
		padding: 1px 5px;
		border-radius: 4px;
	}

	.narrative-body :global(blockquote) {
		border-left: 2px solid var(--color-border);
		padding-left: 14px;
		margin: 12px 0;
		color: var(--color-text-muted);
	}

	.narrative-body :global(a) {
		color: var(--color-accent);
		text-decoration: underline;
	}

	/* ── Amount colors ────────────────────────────────── */
	.income {
		color: var(--color-income);
	}

	.expense {
		color: var(--color-expense);
	}
</style>
