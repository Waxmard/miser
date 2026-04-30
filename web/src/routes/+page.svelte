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

	onMount(async () => {
		try {
			[trends, recentTxns, report, categories, merchantIconOverrides] = await Promise.all([
				api.trends(),
				api.transactions({ limit: 10 }),
				api.latestReport(),
				api.categories(),
				api.merchantIcons()
			]);
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

	$: topCategories = trends?.categories.slice(0, 5) ?? [];
	function sumExpenses(amounts: number[]) {
		return amounts.filter((a) => a < 0).reduce((sum, a) => sum + Math.abs(a), 0);
	}

	$: heroTotal = trends ? sumExpenses(trends.categories.map((c) => c.current)) : 0;
	$: prevMTDTotal = trends ? sumExpenses(trends.categories.map((c) => c.previous)) : 0;
	$: momDeltaPct =
		trends && prevMTDTotal > 0 ? ((heroTotal - prevMTDTotal) / prevMTDTotal) * 100 : null;
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
					vs same point last month (${prevMTDTotal.toLocaleString('en-US', { maximumFractionDigits: 0 })})
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
						{@const pct = cat.budget_used_pct ?? null}
						{@const barPct = pct !== null ? Math.min(100, pct) : null}
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
								<span class="cat-amount {amountClass(cat.current)}">{formatAmount(cat.current)}</span>
							</div>
							{#if barPct !== null && pct !== null}
								<div class="budget-bar-row">
									<div class="budget-bar">
										<div
											class="budget-fill"
											class:over={pct >= 100}
											style="width: {barPct}%"
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
			<section class="report-header">
				<h2>Monthly Report — {report.year}/{String(report.month).padStart(2, '0')}</h2>
			</section>
			{#if report.sections && report.sections.length > 0}
				<div class="report-sections">
					{#each report.sections as section}
						{#if section.type === 'stat'}
							<div class="card report-stat-card">
								<div class="stat-label">{section.title.toUpperCase()}</div>
								<div class="report-stat-value">{section.value}</div>
								{#if section.delta}
									<div class="report-stat-row">
										<span class="report-stat-delta {section.sign === 'positive' ? 'income' : section.sign === 'negative' ? 'expense' : 'muted'}">
											{section.delta}
										</span>
										{#if section.note}
											<span class="report-stat-note">{section.note}</span>
										{/if}
									</div>
								{:else if section.note}
									<div class="report-stat-note">{section.note}</div>
								{/if}
							</div>
						{:else if section.type === 'scorecard'}
							<div class="card">
								<h2>{section.title}</h2>
								<ul class="report-item-list">
									{#each section.items ?? [] as item}
										{@const pct = item.pct ?? 0}
										<li class="report-item">
											<div class="report-item-row">
												<span class="report-item-label">{item.label}</span>
												<span class="report-item-value {item.sign === 'positive' ? 'income' : item.sign === 'negative' ? 'expense' : ''}">{item.value}</span>
											</div>
											{#if item.pct !== undefined}
												<div class="budget-bar-row">
													<div class="budget-bar">
														<div class="budget-fill" class:over={pct >= 100} style="width: {Math.min(100, pct)}%"></div>
													</div>
													<span class="budget-pct" class:over={pct >= 100}>{pct.toFixed(0)}%</span>
												</div>
											{/if}
											{#if item.note}
												<div class="report-item-note">{item.note}</div>
											{/if}
										</li>
									{/each}
								</ul>
							</div>
						{:else if section.type === 'movers' || section.type === 'transactions'}
							<div class="card">
								<h2>{section.title}</h2>
								<ul class="report-item-list">
									{#each section.items ?? [] as item}
										<li class="report-item">
											<div class="report-item-row">
												<span class="report-item-label">{item.label}</span>
												<span class="report-item-value {item.sign === 'positive' ? 'income' : item.sign === 'negative' ? 'expense' : ''}">{item.value ?? item.delta ?? ''}</span>
											</div>
											{#if item.note}
												<div class="report-item-note">{item.note}</div>
											{/if}
										</li>
									{/each}
								</ul>
							</div>
						{:else if section.type === 'takeaways'}
							<div class="card report-takeaways">
								<h2>{section.title}</h2>
								<ol class="takeaway-list">
									{#each section.items ?? [] as item, i}
										<li class="takeaway-item">
											<span class="takeaway-num">{i + 1}</span>
											<span class="takeaway-text">{item.label}</span>
										</li>
									{/each}
								</ol>
							</div>
						{/if}
					{/each}
				</div>
			{:else if report.narrative}
				<article class="narrative">
					<div class="narrative-body">
						{@html marked.parse(report.narrative)}
					</div>
				</article>
			{/if}
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

	/* ── Report sections (structured, below grid) ──────── */
	.report-header {
		margin-top: 32px;
		margin-bottom: 16px;
	}

	.report-sections {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 24px;
	}

	.report-takeaways {
		grid-column: 1 / -1;
	}

	/* stat card */
	.report-stat-card {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.report-stat-value {
		font-family: var(--font-mono);
		font-size: 2rem;
		font-weight: 600;
		color: var(--color-text);
		line-height: 1.1;
		margin: 4px 0;
	}

	.report-stat-row {
		display: flex;
		align-items: baseline;
		gap: 8px;
		flex-wrap: wrap;
	}

	.report-stat-delta {
		font-family: var(--font-mono);
		font-size: 0.882rem;
		font-weight: 500;
	}

	.report-stat-note {
		font-size: 0.824rem;
		color: var(--color-text-muted);
	}

	.muted {
		color: var(--color-text-muted);
	}

	/* item lists (scorecard, movers, transactions) */
	.report-item-list {
		list-style: none;
		display: flex;
		flex-direction: column;
	}

	.report-item {
		padding: 10px 0;
		border-bottom: 1px solid var(--color-border);
	}

	.report-item:last-child {
		border-bottom: none;
	}

	.report-item-row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 8px;
	}

	.report-item-label {
		font-size: 0.824rem;
		font-weight: 500;
		color: var(--color-text);
		flex: 1;
		min-width: 0;
	}

	.report-item-value {
		font-family: var(--font-mono);
		font-size: 0.824rem;
		white-space: nowrap;
	}

	.report-item-note {
		font-size: 0.706rem;
		color: var(--color-text-muted);
		margin-top: 3px;
	}

	/* takeaways */
	.takeaway-list {
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.takeaway-item {
		display: flex;
		align-items: baseline;
		gap: 12px;
	}

	.takeaway-num {
		font-family: var(--font-mono);
		font-size: 0.706rem;
		font-weight: 600;
		color: var(--color-accent);
		min-width: 14px;
		flex-shrink: 0;
	}

	.takeaway-text {
		font-size: 0.882rem;
		color: var(--color-text);
		line-height: 1.5;
	}

	/* ── Narrative fallback (old reports) ───────────────── */
	.narrative {
		margin-top: 8px;
	}

	.narrative-body {
		border-left: 2px solid var(--color-accent);
		padding-left: 24px;
		font-family: var(--font-display);
		font-size: 1rem;
		line-height: 1.75;
		color: var(--color-text-muted);
	}

	.narrative-body :global(p) { margin-bottom: 12px; }
	.narrative-body :global(p:last-child) { margin-bottom: 0; }
	.narrative-body :global(strong) { font-weight: 600; color: var(--color-text); }
	.narrative-body :global(em) { font-style: italic; }

	/* ── Amount colors ────────────────────────────────── */
	.income {
		color: var(--color-income);
	}

	.expense {
		color: var(--color-expense);
	}
</style>
