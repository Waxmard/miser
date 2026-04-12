export type EmojiIcon = {
	emoji: string;
	name: string;
	keywords: string[];
};

/** Parse a stored icon slug into its library and name.
 *  Unnamespaced slugs → simple-icons (backward compat). */
export function parseIconSlug(slug: string): { library: 'si' | 'emoji'; name: string } {
	if (slug.startsWith('emoji:')) return { library: 'emoji', name: slug.slice(6) };
	return { library: 'si', name: slug };
}

export const allEmoji: EmojiIcon[] = [
	// Food & Dining
	{ emoji: '🍕', name: 'Pizza', keywords: ['pizza', 'food', 'italian', 'dining', 'restaurant'] },
	{ emoji: '🍔', name: 'Burger', keywords: ['burger', 'hamburger', 'fast food', 'dining'] },
	{ emoji: '🌮', name: 'Taco', keywords: ['taco', 'mexican', 'food', 'dining'] },
	{ emoji: '🍣', name: 'Sushi', keywords: ['sushi', 'japanese', 'seafood', 'dining', 'restaurant'] },
	{ emoji: '🍜', name: 'Noodles', keywords: ['noodles', 'ramen', 'pasta', 'chinese', 'food', 'dining'] },
	{ emoji: '🥗', name: 'Salad', keywords: ['salad', 'healthy', 'food', 'vegetables', 'dining'] },
	{ emoji: '🍱', name: 'Lunch', keywords: ['lunch', 'bento', 'food', 'meal', 'dining'] },
	{ emoji: '🥩', name: 'Meat', keywords: ['meat', 'steak', 'beef', 'food', 'dining', 'groceries'] },
	{ emoji: '🍞', name: 'Bread', keywords: ['bread', 'bakery', 'food', 'groceries'] },
	{ emoji: '🧁', name: 'Cupcake', keywords: ['cupcake', 'cake', 'dessert', 'bakery', 'sweet'] },
	{ emoji: '🍰', name: 'Cake', keywords: ['cake', 'dessert', 'bakery', 'sweet', 'celebration'] },
	{ emoji: '☕', name: 'Coffee', keywords: ['coffee', 'cafe', 'starbucks', 'drink', 'dining'] },
	{ emoji: '🍺', name: 'Beer', keywords: ['beer', 'bar', 'alcohol', 'drink', 'dining'] },
	{ emoji: '🍷', name: 'Wine', keywords: ['wine', 'bar', 'alcohol', 'drink', 'dining', 'restaurant'] },
	{ emoji: '🍸', name: 'Cocktail', keywords: ['cocktail', 'bar', 'alcohol', 'drink', 'dining'] },
	{ emoji: '🥤', name: 'Drink', keywords: ['drink', 'soda', 'juice', 'beverage', 'boba'] },
	{ emoji: '🧃', name: 'Juice', keywords: ['juice', 'drink', 'smoothie', 'healthy', 'food'] },

	// Groceries
	{ emoji: '🛒', name: 'Groceries', keywords: ['groceries', 'supermarket', 'walmart', 'whole foods', 'shopping', 'food'] },
	{ emoji: '🥦', name: 'Vegetables', keywords: ['vegetables', 'produce', 'groceries', 'healthy', 'food'] },
	{ emoji: '🍎', name: 'Fruit', keywords: ['fruit', 'apple', 'groceries', 'healthy', 'food'] },

	// Transport
	{ emoji: '🚗', name: 'Car', keywords: ['car', 'auto', 'vehicle', 'transport', 'gas', 'parking', 'driving'] },
	{ emoji: '⛽', name: 'Gas', keywords: ['gas', 'fuel', 'petrol', 'car', 'transport'] },
	{ emoji: '🅿️', name: 'Parking', keywords: ['parking', 'car', 'garage', 'transport'] },
	{ emoji: '🚕', name: 'Taxi', keywords: ['taxi', 'uber', 'lyft', 'rideshare', 'transport', 'cab'] },
	{ emoji: '✈️', name: 'Flight', keywords: ['flight', 'plane', 'airplane', 'travel', 'airline', 'transport'] },
	{ emoji: '🚆', name: 'Train', keywords: ['train', 'rail', 'subway', 'metro', 'transit', 'transport'] },
	{ emoji: '🚌', name: 'Bus', keywords: ['bus', 'transit', 'public transport', 'transport'] },
	{ emoji: '🚲', name: 'Bike', keywords: ['bike', 'bicycle', 'cycling', 'transport'] },
	{ emoji: '🛵', name: 'Scooter', keywords: ['scooter', 'moped', 'transport', 'rideshare'] },
	{ emoji: '🚢', name: 'Cruise', keywords: ['cruise', 'ship', 'boat', 'travel', 'vacation'] },

	// Housing
	{ emoji: '🏠', name: 'Home', keywords: ['home', 'house', 'rent', 'mortgage', 'housing', 'apartment'] },
	{ emoji: '🏢', name: 'Office', keywords: ['office', 'work', 'business', 'coworking', 'rent'] },
	{ emoji: '🔑', name: 'Keys', keywords: ['keys', 'rent', 'housing', 'home', 'apartment', 'lease'] },
	{ emoji: '🛋️', name: 'Furniture', keywords: ['furniture', 'ikea', 'home', 'decor', 'interior'] },
	{ emoji: '🧹', name: 'Cleaning', keywords: ['cleaning', 'housekeeping', 'home', 'laundry', 'maintenance'] },
	{ emoji: '🔧', name: 'Repairs', keywords: ['repairs', 'maintenance', 'plumber', 'home', 'tools', 'fix'] },

	// Utilities
	{ emoji: '⚡', name: 'Electricity', keywords: ['electricity', 'electric', 'power', 'utilities', 'energy'] },
	{ emoji: '💧', name: 'Water', keywords: ['water', 'utilities', 'bill', 'sewer'] },
	{ emoji: '🔥', name: 'Gas/Heat', keywords: ['gas', 'heat', 'heating', 'utilities', 'energy', 'natural gas'] },
	{ emoji: '🌐', name: 'Internet', keywords: ['internet', 'wifi', 'broadband', 'utilities', 'telecom', 'isp'] },
	{ emoji: '📱', name: 'Phone', keywords: ['phone', 'mobile', 'cell', 'telecom', 'utilities', 'verizon', 'att', 'tmobile'] },

	// Health
	{ emoji: '💊', name: 'Pharmacy', keywords: ['pharmacy', 'medicine', 'drugs', 'cvs', 'walgreens', 'health', 'prescription'] },
	{ emoji: '🏥', name: 'Medical', keywords: ['medical', 'hospital', 'doctor', 'health', 'healthcare', 'urgent care'] },
	{ emoji: '🦷', name: 'Dental', keywords: ['dental', 'dentist', 'teeth', 'health', 'orthodontist'] },
	{ emoji: '👁️', name: 'Vision', keywords: ['vision', 'eye', 'glasses', 'contacts', 'health', 'optometrist'] },
	{ emoji: '🏋️', name: 'Gym', keywords: ['gym', 'fitness', 'workout', 'health', 'exercise', 'crossfit', 'peloton'] },
	{ emoji: '🧘', name: 'Wellness', keywords: ['wellness', 'yoga', 'meditation', 'spa', 'health', 'massage'] },
	{ emoji: '🩺', name: 'Doctor', keywords: ['doctor', 'physician', 'medical', 'health', 'copay', 'visit'] },
	{ emoji: '🧠', name: 'Mental Health', keywords: ['mental health', 'therapy', 'counseling', 'health', 'therapist'] },

	// Shopping
	{ emoji: '👕', name: 'Clothing', keywords: ['clothing', 'clothes', 'fashion', 'shopping', 'apparel', 'shirt', 'zara', 'h&m'] },
	{ emoji: '👟', name: 'Shoes', keywords: ['shoes', 'sneakers', 'footwear', 'shopping', 'nike', 'adidas'] },
	{ emoji: '👜', name: 'Bag', keywords: ['bag', 'purse', 'handbag', 'shopping', 'fashion', 'accessories'] },
	{ emoji: '💄', name: 'Beauty', keywords: ['beauty', 'makeup', 'cosmetics', 'shopping', 'sephora', 'ulta'] },
	{ emoji: '🛍️', name: 'Shopping', keywords: ['shopping', 'retail', 'store', 'mall', 'purchases'] },
	{ emoji: '📦', name: 'Delivery', keywords: ['delivery', 'amazon', 'shipping', 'package', 'online shopping'] },
	{ emoji: '⌚', name: 'Watch', keywords: ['watch', 'jewelry', 'accessories', 'shopping', 'luxury'] },
	{ emoji: '💍', name: 'Jewelry', keywords: ['jewelry', 'ring', 'accessories', 'shopping', 'luxury'] },

	// Electronics
	{ emoji: '💻', name: 'Computer', keywords: ['computer', 'laptop', 'tech', 'electronics', 'software', 'apple', 'dell'] },
	{ emoji: '📺', name: 'TV', keywords: ['tv', 'television', 'streaming', 'netflix', 'electronics', 'display'] },
	{ emoji: '🎧', name: 'Headphones', keywords: ['headphones', 'audio', 'electronics', 'music', 'airpods', 'earbuds'] },
	{ emoji: '🔋', name: 'Electronics', keywords: ['electronics', 'gadgets', 'tech', 'battery', 'devices', 'accessories'] },
	{ emoji: '📷', name: 'Camera', keywords: ['camera', 'photography', 'electronics', 'photo', 'lens'] },

	// Entertainment
	{ emoji: '🎬', name: 'Movies', keywords: ['movies', 'cinema', 'theater', 'entertainment', 'film', 'amc'] },
	{ emoji: '🎮', name: 'Gaming', keywords: ['gaming', 'games', 'xbox', 'playstation', 'nintendo', 'entertainment', 'steam'] },
	{ emoji: '🎵', name: 'Music', keywords: ['music', 'spotify', 'concerts', 'entertainment', 'streaming', 'apple music'] },
	{ emoji: '📚', name: 'Books', keywords: ['books', 'kindle', 'reading', 'education', 'amazon', 'library', 'audible'] },
	{ emoji: '🎭', name: 'Theater', keywords: ['theater', 'shows', 'broadway', 'entertainment', 'arts', 'performance'] },
	{ emoji: '🎨', name: 'Arts', keywords: ['arts', 'art', 'museum', 'entertainment', 'culture', 'gallery'] },
	{ emoji: '🎡', name: 'Amusement', keywords: ['amusement', 'theme park', 'disneyland', 'entertainment', 'fun', 'tickets'] },

	// Finance
	{ emoji: '💰', name: 'Money', keywords: ['money', 'cash', 'income', 'earnings', 'finance', 'salary'] },
	{ emoji: '💳', name: 'Credit Card', keywords: ['credit card', 'credit', 'payment', 'finance', 'card', 'debit'] },
	{ emoji: '🏦', name: 'Bank', keywords: ['bank', 'banking', 'finance', 'savings', 'account', 'chase', 'wells fargo'] },
	{ emoji: '💵', name: 'Cash', keywords: ['cash', 'atm', 'withdrawal', 'finance', 'money', 'bills'] },
	{ emoji: '📈', name: 'Investments', keywords: ['investments', 'stocks', 'investing', 'portfolio', 'finance', 'schwab', 'fidelity', 'trading'] },
	{ emoji: '💸', name: 'Transfer', keywords: ['transfer', 'send money', 'venmo', 'paypal', 'zelle', 'finance', 'wire'] },
	{ emoji: '🪙', name: 'Fees', keywords: ['fees', 'coins', 'change', 'finance', 'charges', 'service fee'] },
	{ emoji: '🧾', name: 'Receipt', keywords: ['receipt', 'bill', 'invoice', 'finance', 'payment', 'expense'] },
	{ emoji: '📊', name: 'Budget', keywords: ['budget', 'finance', 'expenses', 'tracking', 'spending'] },
	{ emoji: '🏧', name: 'ATM', keywords: ['atm', 'cash', 'withdrawal', 'bank', 'finance'] },

	// Travel
	{ emoji: '🏨', name: 'Hotel', keywords: ['hotel', 'lodging', 'accommodation', 'travel', 'airbnb', 'marriott', 'hilton'] },
	{ emoji: '🗺️', name: 'Travel', keywords: ['travel', 'vacation', 'trip', 'holiday', 'tourism'] },
	{ emoji: '🏖️', name: 'Beach', keywords: ['beach', 'vacation', 'travel', 'resort', 'tropical'] },
	{ emoji: '⛷️', name: 'Skiing', keywords: ['skiing', 'ski', 'snowboard', 'winter', 'travel', 'vacation', 'mountain'] },
	{ emoji: '🧳', name: 'Luggage', keywords: ['luggage', 'travel', 'trip', 'vacation', 'bags', 'suitcase'] },

	// Education
	{ emoji: '🎓', name: 'Education', keywords: ['education', 'school', 'tuition', 'college', 'university', 'student', 'loans'] },
	{ emoji: '✏️', name: 'School Supplies', keywords: ['school', 'supplies', 'stationery', 'education', 'office'] },
	{ emoji: '🧑‍💻', name: 'Online Learning', keywords: ['online course', 'udemy', 'coursera', 'education', 'training', 'learning', 'udemy'] },

	// Personal Care
	{ emoji: '💈', name: 'Barber', keywords: ['barber', 'haircut', 'salon', 'grooming', 'personal care', 'hair'] },
	{ emoji: '💅', name: 'Nails', keywords: ['nails', 'manicure', 'pedicure', 'salon', 'beauty', 'personal care'] },
	{ emoji: '🧴', name: 'Skincare', keywords: ['skincare', 'beauty', 'personal care', 'cosmetics', 'lotion', 'moisturizer'] },
	{ emoji: '🪒', name: 'Grooming', keywords: ['grooming', 'shaving', 'personal care', 'razor', 'barber'] },

	// Pets
	{ emoji: '🐕', name: 'Dog', keywords: ['dog', 'pet', 'vet', 'animal', 'petco', 'petsmart', 'grooming'] },
	{ emoji: '🐈', name: 'Cat', keywords: ['cat', 'pet', 'vet', 'animal', 'petco', 'petsmart'] },
	{ emoji: '🐾', name: 'Pets', keywords: ['pets', 'vet', 'veterinary', 'animal', 'petco', 'petsmart', 'pet supplies'] },

	// Kids & Family
	{ emoji: '👶', name: 'Baby', keywords: ['baby', 'infant', 'diapers', 'childcare', 'kids', 'formula'] },
	{ emoji: '🎒', name: 'School', keywords: ['school', 'backpack', 'kids', 'education', 'children', 'supplies'] },
	{ emoji: '🧸', name: 'Toys', keywords: ['toys', 'kids', 'children', 'games', 'play', 'lego'] },

	// Home Improvement
	{ emoji: '🏗️', name: 'Construction', keywords: ['construction', 'renovation', 'remodel', 'home improvement', 'contractor'] },
	{ emoji: '🪴', name: 'Plants', keywords: ['plants', 'garden', 'home', 'gardening', 'nursery', 'flowers'] },
	{ emoji: '🪣', name: 'Hardware', keywords: ['hardware', 'home supplies', 'home depot', 'lowes', 'tools', 'diy'] },

	// Subscriptions & Streaming
	{ emoji: '🔄', name: 'Subscription', keywords: ['subscription', 'recurring', 'monthly', 'annual', 'membership', 'saas'] },
	{ emoji: '📡', name: 'Streaming', keywords: ['streaming', 'netflix', 'hulu', 'disney', 'subscription', 'entertainment', 'cable'] },

	// Insurance & Taxes
	{ emoji: '🏛️', name: 'Taxes', keywords: ['taxes', 'irs', 'government', 'finance', 'tax', 'return'] },
	{ emoji: '🛡️', name: 'Insurance', keywords: ['insurance', 'health insurance', 'car insurance', 'home insurance', 'finance', 'premium'] },

	// Gifts & Charity
	{ emoji: '🎁', name: 'Gifts', keywords: ['gifts', 'presents', 'birthday', 'holiday', 'shopping', 'amazon'] },
	{ emoji: '❤️', name: 'Charity', keywords: ['charity', 'donation', 'nonprofit', 'giving', 'volunteer'] },

	// Work & Business
	{ emoji: '💼', name: 'Business', keywords: ['business', 'work', 'professional', 'office', 'corporate', 'b2b'] },
	{ emoji: '🖨️', name: 'Office Supplies', keywords: ['office supplies', 'staples', 'work', 'printer', 'paper'] },

	// Sports & Recreation
	{ emoji: '⚽', name: 'Sports', keywords: ['sports', 'soccer', 'recreation', 'fitness', 'league'] },
	{ emoji: '🎾', name: 'Tennis', keywords: ['tennis', 'sports', 'recreation', 'fitness', 'court'] },
	{ emoji: '🏊', name: 'Swimming', keywords: ['swimming', 'pool', 'fitness', 'sports', 'recreation', 'aquatics'] },
	{ emoji: '🧗', name: 'Climbing', keywords: ['climbing', 'rock climbing', 'fitness', 'sports', 'recreation', 'bouldering'] },

	// Events & Misc
	{ emoji: '🎉', name: 'Events', keywords: ['events', 'party', 'celebration', 'entertainment', 'tickets', 'concert'] },
	{ emoji: '📮', name: 'Mail', keywords: ['mail', 'postage', 'stamps', 'shipping', 'usps', 'fedex', 'ups'] },
	{ emoji: '🌱', name: 'Garden', keywords: ['garden', 'plants', 'outdoors', 'home', 'nursery', 'seeds'] },
	{ emoji: '☀️', name: 'Outdoor', keywords: ['outdoor', 'activities', 'recreation', 'park', 'hiking', 'camping'] },
	{ emoji: '🎰', name: 'Gambling', keywords: ['gambling', 'casino', 'lottery', 'betting', 'poker'] },
	{ emoji: '🚬', name: 'Tobacco', keywords: ['tobacco', 'cigarettes', 'smoking', 'vaping'] },
	{ emoji: '🍃', name: 'Cannabis', keywords: ['cannabis', 'dispensary', 'marijuana', 'weed'] },
];
