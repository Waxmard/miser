const BASE = '/api';

export interface Transaction {
	id: string;
	account_id: string;
	account_name: string;
	category_id: string | null;
	category_name: string;
	amount: number;
	merchant: string;
	merchant_clean?: string;
	description?: string;
	date: string;
	source: string;
	status: string;
	tags?: string;
	owner?: string;
	notes?: string;
}

export interface Category {
	id: string;
	name: string;
	parent_id: string | null;
	icon?: string;
	transaction_count: number;
	total_amount: number;
}

export interface CategoryTotal {
	category: string;
	total: number;
	count: number;
	subcategories?: CategoryTotal[];
}

export interface TrendsResponse {
	current_month: string;
	previous_month: string;
	current: CategoryTotal[];
	previous: CategoryTotal[];
	budgets: { category: string; budget: number }[];
}

export interface Budget {
	id: string;
	category_id: string;
	category_name: string;
	monthly_amount: number;
}

export interface Account {
	id: string;
	name: string;
	institution: string;
	account_type: string;
	source: string;
}

export interface Report {
	id: string;
	year: number;
	month: number;
	narrative: string;
}

export interface MerchantIcon {
	merchant_name: string;
	icon_slug: string;
	updated_at: string;
}

export interface TransactionFilters {
	from?: string;
	to?: string;
	category?: string;
	account?: string;
	tag?: string;
	owner?: string;
	q?: string;
	limit?: number;
	offset?: number;
}

async function get<T>(path: string, params?: Record<string, string>): Promise<T> {
	const url = new URL(BASE + path, globalThis.location.origin);
	if (params) {
		for (const [k, v] of Object.entries(params)) {
			if (v !== undefined && v !== '') url.searchParams.set(k, v);
		}
	}
	const res = await fetch(url.toString());
	if (!res.ok) throw new Error(`${path}: ${res.status}`);
	return res.json() as Promise<T>;
}

async function patch<T>(path: string, body: unknown): Promise<T> {
	const res = await fetch(BASE + path, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	if (!res.ok) throw new Error(`${path}: ${res.status}`);
	return res.json() as Promise<T>;
}

async function put<T>(path: string, body: unknown): Promise<T> {
	const res = await fetch(BASE + path, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	if (!res.ok) throw new Error(`${path}: ${res.status}`);
	return res.json() as Promise<T>;
}

async function del(path: string): Promise<void> {
	const res = await fetch(BASE + path, { method: 'DELETE' });
	if (!res.ok) throw new Error(`${path}: ${res.status}`);
}

export const api = {
	transactions: (filters: TransactionFilters = {}) =>
		get<Transaction[]>('/transactions', filters as Record<string, string>),

	categories: () => get<Category[]>('/categories'),

	updateCategoryIcon: (id: string, icon: string | null) =>
		patch<{ id: string; icon: string | null }>(`/categories/${id}`, { icon }),

	trends: () => get<TrendsResponse>('/trends'),

	budgets: () => get<Budget[]>('/budgets'),

	accounts: () => get<Account[]>('/accounts'),

	latestReport: () => get<Report | null>('/reports/latest'),

	merchantIcons: () => get<MerchantIcon[]>('/merchant-icons'),

	setMerchantIcon: (merchant_name: string, icon_slug: string) =>
		put<MerchantIcon>('/merchant-icons', { merchant_name, icon_slug }),

	deleteMerchantIcon: (merchantName: string) =>
		del(`/merchant-icons?name=${encodeURIComponent(merchantName)}`)
};
