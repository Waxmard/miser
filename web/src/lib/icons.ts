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
	{ emoji: '🌮', name: 'Taco', keywords: ['taco', 'mexican', 'food', 'dining', 'chipotle'] },
	{ emoji: '🍣', name: 'Sushi', keywords: ['sushi', 'japanese', 'seafood', 'dining', 'restaurant'] },
	{ emoji: '🍜', name: 'Noodles', keywords: ['noodles', 'ramen', 'pasta', 'chinese', 'food', 'dining', 'pho'] },
	{ emoji: '🥗', name: 'Salad', keywords: ['salad', 'healthy', 'food', 'vegetables', 'dining', 'panera'] },
	{ emoji: '🍱', name: 'Lunch', keywords: ['lunch', 'bento', 'food', 'meal', 'dining'] },
	{ emoji: '🥩', name: 'Meat', keywords: ['meat', 'steak', 'beef', 'food', 'dining', 'groceries', 'steakhouse'] },
	{ emoji: '🍞', name: 'Bread', keywords: ['bread', 'bakery', 'food', 'groceries'] },
	{ emoji: '🥐', name: 'Croissant', keywords: ['croissant', 'pastry', 'bakery', 'breakfast', 'french', 'cafe'] },
	{ emoji: '🧇', name: 'Waffle', keywords: ['waffle', 'pancake', 'breakfast', 'brunch', 'ihop', 'dining'] },
	{ emoji: '🥪', name: 'Sandwich', keywords: ['sandwich', 'sub', 'deli', 'lunch', 'dining', 'subway', 'jersey mike'] },
	{ emoji: '🍗', name: 'Chicken', keywords: ['chicken', 'wings', 'poultry', 'fast food', 'dining', 'chick-fil-a', 'kfc'] },
	{ emoji: '🌯', name: 'Wrap', keywords: ['wrap', 'burrito', 'mexican', 'food', 'dining', 'chipotle', 'qdoba'] },
	{ emoji: '🍩', name: 'Donut', keywords: ['donut', 'doughnut', 'pastry', 'breakfast', 'dunkin', 'sweet'] },
	{ emoji: '🍦', name: 'Ice Cream', keywords: ['ice cream', 'gelato', 'dessert', 'sweet', 'dairy queen', 'baskin robbins'] },
	{ emoji: '🍫', name: 'Chocolate', keywords: ['chocolate', 'candy', 'sweet', 'dessert', 'snack', 'food'] },
	{ emoji: '🧁', name: 'Cupcake', keywords: ['cupcake', 'cake', 'dessert', 'bakery', 'sweet'] },
	{ emoji: '🍰', name: 'Cake', keywords: ['cake', 'dessert', 'bakery', 'sweet', 'celebration', 'cheesecake factory'] },
	{ emoji: '☕', name: 'Coffee', keywords: ['coffee', 'cafe', 'starbucks', 'dunkin', 'drink', 'dining', 'espresso', 'latte'] },
	{ emoji: '🫖', name: 'Tea', keywords: ['tea', 'drink', 'herbal', 'cafe', 'boba', 'chai'] },
	{ emoji: '🧋', name: 'Bubble Tea', keywords: ['bubble tea', 'boba', 'drink', 'tea', 'cafe'] },
	{ emoji: '🍺', name: 'Beer', keywords: ['beer', 'bar', 'alcohol', 'drink', 'dining', 'brewery', 'pub'] },
	{ emoji: '🍷', name: 'Wine', keywords: ['wine', 'bar', 'alcohol', 'drink', 'dining', 'restaurant', 'winery'] },
	{ emoji: '🥃', name: 'Spirits', keywords: ['whiskey', 'bourbon', 'spirits', 'liquor', 'bar', 'alcohol', 'drink'] },
	{ emoji: '🍸', name: 'Cocktail', keywords: ['cocktail', 'bar', 'alcohol', 'drink', 'dining', 'mixology'] },
	{ emoji: '🥤', name: 'Drink', keywords: ['drink', 'soda', 'juice', 'beverage', 'boba', 'smoothie'] },
	{ emoji: '🧃', name: 'Juice', keywords: ['juice', 'drink', 'smoothie', 'healthy', 'food', 'pressed'] },

	// Groceries
	{ emoji: '🛒', name: 'Groceries', keywords: ['groceries', 'supermarket', 'walmart', 'whole foods', 'kroger', 'safeway', 'trader joe', 'aldi', 'publix', 'shopping', 'food'] },
	{ emoji: '🥦', name: 'Vegetables', keywords: ['vegetables', 'produce', 'groceries', 'healthy', 'food', 'farmers market'] },
	{ emoji: '🍎', name: 'Fruit', keywords: ['fruit', 'apple', 'groceries', 'healthy', 'food', 'produce'] },
	{ emoji: '🥑', name: 'Avocado', keywords: ['avocado', 'produce', 'groceries', 'healthy', 'food'] },
	{ emoji: '🧀', name: 'Dairy', keywords: ['dairy', 'cheese', 'milk', 'yogurt', 'groceries', 'food'] },
	{ emoji: '🫐', name: 'Berries', keywords: ['berries', 'blueberry', 'fruit', 'groceries', 'healthy', 'food'] },

	// Transport
	{ emoji: '🚗', name: 'Car', keywords: ['car', 'auto', 'vehicle', 'transport', 'gas', 'parking', 'driving'] },
	{ emoji: '⛽', name: 'Gas', keywords: ['gas', 'fuel', 'petrol', 'car', 'transport', 'shell', 'chevron', 'bp', 'exxon'] },
	{ emoji: '🅿️', name: 'Parking', keywords: ['parking', 'car', 'garage', 'transport', 'meter'] },
	{ emoji: '🚕', name: 'Taxi', keywords: ['taxi', 'uber', 'lyft', 'rideshare', 'transport', 'cab'] },
	{ emoji: '✈️', name: 'Flight', keywords: ['flight', 'plane', 'airplane', 'travel', 'airline', 'transport', 'airport'] },
	{ emoji: '🚆', name: 'Train', keywords: ['train', 'rail', 'subway', 'metro', 'transit', 'transport', 'amtrak'] },
	{ emoji: '🚌', name: 'Bus', keywords: ['bus', 'transit', 'public transport', 'transport', 'coach'] },
	{ emoji: '🚲', name: 'Bike', keywords: ['bike', 'bicycle', 'cycling', 'transport', 'citibike', 'lime', 'bird'] },
	{ emoji: '🛵', name: 'Scooter', keywords: ['scooter', 'moped', 'transport', 'rideshare', 'lime', 'bird'] },
	{ emoji: '🏍️', name: 'Motorcycle', keywords: ['motorcycle', 'motorbike', 'harley', 'transport', 'insurance'] },
	{ emoji: '🚢', name: 'Cruise', keywords: ['cruise', 'ship', 'boat', 'travel', 'vacation', 'carnival', 'royal caribbean'] },
	{ emoji: '🚁', name: 'Helicopter', keywords: ['helicopter', 'charter', 'transport', 'travel', 'tour'] },
	{ emoji: '🛻', name: 'Truck', keywords: ['truck', 'pickup', 'moving', 'transport', 'uhaul', 'penske'] },
	{ emoji: '🚐', name: 'Van', keywords: ['van', 'shuttle', 'transport', 'moving', 'rental'] },
	{ emoji: '🛳️', name: 'Ferry', keywords: ['ferry', 'boat', 'water', 'transport', 'travel'] },

	// Housing
	{ emoji: '🏠', name: 'Home', keywords: ['home', 'house', 'rent', 'mortgage', 'housing', 'apartment'] },
	{ emoji: '🏢', name: 'Office', keywords: ['office', 'work', 'business', 'coworking', 'rent', 'wework'] },
	{ emoji: '🔑', name: 'Keys', keywords: ['keys', 'rent', 'housing', 'home', 'apartment', 'lease'] },
	{ emoji: '🛋️', name: 'Furniture', keywords: ['furniture', 'ikea', 'home', 'decor', 'interior', 'wayfair', 'west elm'] },
	{ emoji: '🧹', name: 'Cleaning', keywords: ['cleaning', 'housekeeping', 'home', 'laundry', 'maintenance', 'maid', 'detergent'] },
	{ emoji: '🔧', name: 'Repairs', keywords: ['repairs', 'maintenance', 'plumber', 'home', 'tools', 'fix', 'handyman'] },
	{ emoji: '🪟', name: 'Windows', keywords: ['windows', 'blinds', 'curtains', 'home', 'decor', 'renovation'] },
	{ emoji: '🧰', name: 'Toolbox', keywords: ['tools', 'toolbox', 'diy', 'home depot', 'lowes', 'hardware', 'repair'] },
	{ emoji: '🚿', name: 'Plumbing', keywords: ['plumbing', 'shower', 'bathroom', 'home', 'repair', 'maintenance'] },

	// Utilities
	{ emoji: '⚡', name: 'Electricity', keywords: ['electricity', 'electric', 'power', 'utilities', 'energy', 'pge', 'con ed'] },
	{ emoji: '💧', name: 'Water', keywords: ['water', 'utilities', 'bill', 'sewer', 'sewage'] },
	{ emoji: '🔥', name: 'Gas/Heat', keywords: ['gas', 'heat', 'heating', 'utilities', 'energy', 'natural gas', 'national grid'] },
	{ emoji: '🌐', name: 'Internet', keywords: ['internet', 'wifi', 'broadband', 'utilities', 'telecom', 'isp', 'comcast', 'xfinity', 'at&t'] },
	{ emoji: '📱', name: 'Phone', keywords: ['phone', 'mobile', 'cell', 'telecom', 'utilities', 'verizon', 'att', 'tmobile', 'mint', 'visible'] },
	{ emoji: '📺', name: 'Cable', keywords: ['cable', 'tv', 'comcast', 'spectrum', 'utilities', 'satellite', 'directv'] },

	// Health
	{ emoji: '💊', name: 'Pharmacy', keywords: ['pharmacy', 'medicine', 'drugs', 'cvs', 'walgreens', 'rite aid', 'health', 'prescription'] },
	{ emoji: '🏥', name: 'Medical', keywords: ['medical', 'hospital', 'doctor', 'health', 'healthcare', 'urgent care', 'clinic'] },
	{ emoji: '🦷', name: 'Dental', keywords: ['dental', 'dentist', 'teeth', 'health', 'orthodontist', 'braces'] },
	{ emoji: '👁️', name: 'Vision', keywords: ['vision', 'eye', 'glasses', 'contacts', 'health', 'optometrist', 'warby parker'] },
	{ emoji: '🏋️', name: 'Gym', keywords: ['gym', 'fitness', 'workout', 'health', 'exercise', 'crossfit', 'peloton', 'planet fitness', 'anytime fitness'] },
	{ emoji: '🧘', name: 'Wellness', keywords: ['wellness', 'yoga', 'meditation', 'spa', 'health', 'massage', 'acupuncture'] },
	{ emoji: '🩺', name: 'Doctor', keywords: ['doctor', 'physician', 'medical', 'health', 'copay', 'visit', 'primary care'] },
	{ emoji: '🧠', name: 'Mental Health', keywords: ['mental health', 'therapy', 'counseling', 'health', 'therapist', 'betterhelp', 'talkspace'] },
	{ emoji: '🏃', name: 'Running', keywords: ['running', 'jogging', 'marathon', 'fitness', 'race', 'strava'] },
	{ emoji: '💉', name: 'Vaccine', keywords: ['vaccine', 'injection', 'shot', 'health', 'immunization', 'flu shot'] },
	{ emoji: '🩹', name: 'First Aid', keywords: ['first aid', 'bandage', 'health', 'wound', 'care'] },

	// Shopping
	{ emoji: '👕', name: 'Clothing', keywords: ['clothing', 'clothes', 'fashion', 'shopping', 'apparel', 'shirt', 'zara', 'h&m', 'gap', 'old navy', 'uniqlo'] },
	{ emoji: '👟', name: 'Shoes', keywords: ['shoes', 'sneakers', 'footwear', 'shopping', 'nike', 'adidas', 'new balance', 'foot locker'] },
	{ emoji: '👜', name: 'Bag', keywords: ['bag', 'purse', 'handbag', 'shopping', 'fashion', 'accessories', 'coach', 'louis vuitton'] },
	{ emoji: '💄', name: 'Beauty', keywords: ['beauty', 'makeup', 'cosmetics', 'shopping', 'sephora', 'ulta', 'fenty', 'nars'] },
	{ emoji: '🕶️', name: 'Sunglasses', keywords: ['sunglasses', 'glasses', 'eyewear', 'accessories', 'shopping', 'ray-ban', 'oakley'] },
	{ emoji: '🛍️', name: 'Shopping', keywords: ['shopping', 'retail', 'store', 'mall', 'purchases', 'outlet'] },
	{ emoji: '📦', name: 'Delivery', keywords: ['delivery', 'amazon', 'shipping', 'package', 'online shopping', 'ups', 'fedex', 'usps'] },
	{ emoji: '⌚', name: 'Watch', keywords: ['watch', 'jewelry', 'accessories', 'shopping', 'luxury', 'rolex', 'apple watch'] },
	{ emoji: '💍', name: 'Jewelry', keywords: ['jewelry', 'ring', 'accessories', 'shopping', 'luxury', 'tiffany'] },

	// Electronics
	{ emoji: '💻', name: 'Computer', keywords: ['computer', 'laptop', 'tech', 'electronics', 'software', 'apple', 'dell', 'lenovo', 'hp'] },
	{ emoji: '🖥️', name: 'Desktop', keywords: ['desktop', 'monitor', 'computer', 'tech', 'electronics', 'display', 'pc'] },
	{ emoji: '🎧', name: 'Headphones', keywords: ['headphones', 'audio', 'electronics', 'music', 'airpods', 'earbuds', 'sony', 'bose'] },
	{ emoji: '🔋', name: 'Electronics', keywords: ['electronics', 'gadgets', 'tech', 'battery', 'devices', 'accessories', 'best buy'] },
	{ emoji: '📷', name: 'Camera', keywords: ['camera', 'photography', 'electronics', 'photo', 'lens', 'canon', 'nikon', 'sony'] },
	{ emoji: '🖨️', name: 'Printer', keywords: ['printer', 'ink', 'toner', 'office', 'electronics', 'hp', 'epson'] },
	{ emoji: '🎙️', name: 'Microphone', keywords: ['microphone', 'podcast', 'recording', 'audio', 'streaming', 'content'] },

	// Entertainment
	{ emoji: '🎬', name: 'Movies', keywords: ['movies', 'cinema', 'theater', 'entertainment', 'film', 'amc', 'regal', 'fandango'] },
	{ emoji: '🎮', name: 'Gaming', keywords: ['gaming', 'games', 'xbox', 'playstation', 'nintendo', 'entertainment', 'steam', 'epic games'] },
	{ emoji: '🎵', name: 'Music', keywords: ['music', 'spotify', 'concerts', 'entertainment', 'streaming', 'apple music', 'tidal', 'pandora'] },
	{ emoji: '🎸', name: 'Guitar', keywords: ['guitar', 'music', 'instrument', 'entertainment', 'band', 'lessons'] },
	{ emoji: '📚', name: 'Books', keywords: ['books', 'kindle', 'reading', 'education', 'amazon', 'library', 'audible', 'barnes noble'] },
	{ emoji: '🎭', name: 'Theater', keywords: ['theater', 'shows', 'broadway', 'entertainment', 'arts', 'performance', 'opera'] },
	{ emoji: '🎨', name: 'Arts', keywords: ['arts', 'art', 'museum', 'entertainment', 'culture', 'gallery', 'exhibition'] },
	{ emoji: '🎡', name: 'Amusement', keywords: ['amusement', 'theme park', 'disneyland', 'entertainment', 'fun', 'tickets', 'six flags'] },
	{ emoji: '🎲', name: 'Board Games', keywords: ['board games', 'games', 'tabletop', 'entertainment', 'hobby', 'dice'] },
	{ emoji: '🎳', name: 'Bowling', keywords: ['bowling', 'recreation', 'entertainment', 'sports', 'alley'] },
	{ emoji: '🎯', name: 'Darts', keywords: ['darts', 'recreation', 'entertainment', 'sports', 'bar games'] },
	{ emoji: '🎪', name: 'Circus', keywords: ['circus', 'carnival', 'fair', 'entertainment', 'show', 'tickets'] },

	// Finance
	{ emoji: '💰', name: 'Money', keywords: ['money', 'cash', 'income', 'earnings', 'finance', 'salary', 'paycheck'] },
	{ emoji: '💳', name: 'Credit Card', keywords: ['credit card', 'credit', 'payment', 'finance', 'card', 'debit', 'amex', 'visa', 'mastercard'] },
	{ emoji: '🏦', name: 'Bank', keywords: ['bank', 'banking', 'finance', 'savings', 'account', 'chase', 'wells fargo', 'bank of america'] },
	{ emoji: '💵', name: 'Cash', keywords: ['cash', 'atm', 'withdrawal', 'finance', 'money', 'bills'] },
	{ emoji: '📈', name: 'Investments', keywords: ['investments', 'stocks', 'investing', 'portfolio', 'finance', 'schwab', 'fidelity', 'trading', 'robinhood', 'etrade', 'vanguard'] },
	{ emoji: '💸', name: 'Transfer', keywords: ['transfer', 'send money', 'venmo', 'paypal', 'zelle', 'finance', 'wire', 'cashapp'] },
	{ emoji: '🪙', name: 'Fees', keywords: ['fees', 'coins', 'change', 'finance', 'charges', 'service fee', 'annual fee'] },
	{ emoji: '🧾', name: 'Receipt', keywords: ['receipt', 'bill', 'invoice', 'finance', 'payment', 'expense'] },
	{ emoji: '📊', name: 'Budget', keywords: ['budget', 'finance', 'expenses', 'tracking', 'spending', 'planning'] },
	{ emoji: '🏧', name: 'ATM', keywords: ['atm', 'cash', 'withdrawal', 'bank', 'finance'] },
	{ emoji: '🐷', name: 'Savings', keywords: ['savings', 'piggy bank', 'save', 'finance', 'emergency fund', 'nest egg'] },
	{ emoji: '₿', name: 'Crypto', keywords: ['crypto', 'bitcoin', 'ethereum', 'coinbase', 'finance', 'blockchain'] },

	// Travel
	{ emoji: '🏨', name: 'Hotel', keywords: ['hotel', 'lodging', 'accommodation', 'travel', 'airbnb', 'marriott', 'hilton', 'hyatt', 'ihg'] },
	{ emoji: '🗺️', name: 'Travel', keywords: ['travel', 'vacation', 'trip', 'holiday', 'tourism', 'expedia', 'booking'] },
	{ emoji: '🏖️', name: 'Beach', keywords: ['beach', 'vacation', 'travel', 'resort', 'tropical', 'ocean'] },
	{ emoji: '⛷️', name: 'Skiing', keywords: ['skiing', 'ski', 'snowboard', 'winter', 'travel', 'vacation', 'mountain', 'lodge'] },
	{ emoji: '🏕️', name: 'Camping', keywords: ['camping', 'outdoors', 'tent', 'national park', 'travel', 'hiking', 'rei'] },
	{ emoji: '🧳', name: 'Luggage', keywords: ['luggage', 'travel', 'trip', 'vacation', 'bags', 'suitcase', 'away'] },
	{ emoji: '🗼', name: 'Tourism', keywords: ['tourism', 'sightseeing', 'travel', 'landmark', 'tour', 'attraction'] },

	// Education
	{ emoji: '🎓', name: 'Education', keywords: ['education', 'school', 'tuition', 'college', 'university', 'student', 'loans'] },
	{ emoji: '✏️', name: 'School Supplies', keywords: ['school', 'supplies', 'stationery', 'education', 'office', 'staples'] },
	{ emoji: '🧑‍💻', name: 'Online Learning', keywords: ['online course', 'udemy', 'coursera', 'education', 'training', 'learning', 'skillshare', 'masterclass'] },
	{ emoji: '📝', name: 'Notes', keywords: ['notes', 'writing', 'study', 'education', 'journal', 'notability'] },

	// Personal Care
	{ emoji: '💈', name: 'Barber', keywords: ['barber', 'haircut', 'salon', 'grooming', 'personal care', 'hair', 'great clips'] },
	{ emoji: '💅', name: 'Nails', keywords: ['nails', 'manicure', 'pedicure', 'salon', 'beauty', 'personal care'] },
	{ emoji: '🧴', name: 'Skincare', keywords: ['skincare', 'beauty', 'personal care', 'cosmetics', 'lotion', 'moisturizer', 'cerave'] },
	{ emoji: '🪒', name: 'Grooming', keywords: ['grooming', 'shaving', 'personal care', 'razor', 'barber', 'dollar shave'] },
	{ emoji: '🧼', name: 'Soap', keywords: ['soap', 'hygiene', 'personal care', 'body wash', 'hand wash', 'cleaning'] },
	{ emoji: '🪥', name: 'Dental Care', keywords: ['toothbrush', 'toothpaste', 'dental', 'hygiene', 'personal care', 'oral-b'] },
	{ emoji: '🛁', name: 'Spa', keywords: ['spa', 'bath', 'relaxation', 'wellness', 'personal care', 'massage'] },
	{ emoji: '🪞', name: 'Mirror', keywords: ['mirror', 'beauty', 'personal care', 'grooming', 'vanity'] },

	// Pets
	{ emoji: '🐕', name: 'Dog', keywords: ['dog', 'pet', 'vet', 'animal', 'petco', 'petsmart', 'grooming', 'chewy'] },
	{ emoji: '🐈', name: 'Cat', keywords: ['cat', 'pet', 'vet', 'animal', 'petco', 'petsmart', 'chewy', 'litter'] },
	{ emoji: '🐾', name: 'Pets', keywords: ['pets', 'vet', 'veterinary', 'animal', 'petco', 'petsmart', 'pet supplies', 'chewy'] },
	{ emoji: '🦮', name: 'Dog Walking', keywords: ['dog walker', 'pet care', 'rover', 'wag', 'dog sitting', 'pet service'] },

	// Kids & Family
	{ emoji: '👶', name: 'Baby', keywords: ['baby', 'infant', 'diapers', 'childcare', 'kids', 'formula', 'baby gear'] },
	{ emoji: '🎒', name: 'School', keywords: ['school', 'backpack', 'kids', 'education', 'children', 'supplies'] },
	{ emoji: '🧸', name: 'Toys', keywords: ['toys', 'kids', 'children', 'games', 'play', 'lego', 'target', 'amazon'] },
	{ emoji: '🍼', name: 'Baby Care', keywords: ['baby', 'formula', 'bottle', 'childcare', 'infant', 'newborn'] },
	{ emoji: '🎠', name: 'Playground', keywords: ['playground', 'park', 'kids', 'children', 'recreation', 'family'] },

	// Home Improvement
	{ emoji: '🏗️', name: 'Construction', keywords: ['construction', 'renovation', 'remodel', 'home improvement', 'contractor', 'builder'] },
	{ emoji: '🪴', name: 'Plants', keywords: ['plants', 'garden', 'home', 'gardening', 'nursery', 'flowers', 'homedepot'] },
	{ emoji: '🪣', name: 'Hardware', keywords: ['hardware', 'home supplies', 'home depot', 'lowes', 'tools', 'diy', 'menards'] },
	{ emoji: '🌿', name: 'Garden', keywords: ['garden', 'landscaping', 'lawn', 'outdoor', 'home', 'plants', 'seeds', 'mulch'] },
	{ emoji: '🔐', name: 'Security', keywords: ['security', 'alarm', 'camera', 'lock', 'home', 'adt', 'ring', 'simplisafe'] },

	// Subscriptions & Streaming
	{ emoji: '🔄', name: 'Subscription', keywords: ['subscription', 'recurring', 'monthly', 'annual', 'membership', 'saas'] },
	{ emoji: '📡', name: 'Streaming', keywords: ['streaming', 'netflix', 'hulu', 'disney', 'hbo', 'peacock', 'paramount', 'subscription', 'entertainment', 'cable'] },

	// Insurance & Taxes
	{ emoji: '🏛️', name: 'Taxes', keywords: ['taxes', 'irs', 'government', 'finance', 'tax', 'return', 'turbotax', 'h&r block'] },
	{ emoji: '🛡️', name: 'Insurance', keywords: ['insurance', 'health insurance', 'car insurance', 'home insurance', 'life insurance', 'finance', 'premium', 'geico', 'progressive', 'state farm'] },

	// Gifts & Charity
	{ emoji: '🎁', name: 'Gifts', keywords: ['gifts', 'presents', 'birthday', 'holiday', 'shopping', 'amazon', 'etsy'] },
	{ emoji: '❤️', name: 'Charity', keywords: ['charity', 'donation', 'nonprofit', 'giving', 'volunteer', 'patreon', 'gofundme'] },
	{ emoji: '🎀', name: 'Gift Wrap', keywords: ['gift wrap', 'wrapping', 'birthday', 'holiday', 'presents', 'celebration'] },

	// Work & Business
	{ emoji: '💼', name: 'Business', keywords: ['business', 'work', 'professional', 'office', 'corporate', 'b2b', 'consulting'] },
	{ emoji: '📊', name: 'Office Supplies', keywords: ['office supplies', 'staples', 'work', 'printer', 'paper', 'amazon business'] },
	{ emoji: '🖥️', name: 'Software', keywords: ['software', 'saas', 'subscription', 'adobe', 'microsoft', 'slack', 'zoom', 'work'] },
	{ emoji: '🤝', name: 'Contract', keywords: ['contract', 'freelance', 'business', 'consulting', 'deal', 'invoice'] },

	// Sports & Recreation
	{ emoji: '⚽', name: 'Soccer', keywords: ['soccer', 'football', 'sports', 'recreation', 'fitness', 'league'] },
	{ emoji: '🏈', name: 'Football', keywords: ['football', 'nfl', 'sports', 'recreation', 'tickets', 'tailgate'] },
	{ emoji: '🏀', name: 'Basketball', keywords: ['basketball', 'nba', 'sports', 'recreation', 'court', 'tickets'] },
	{ emoji: '⚾', name: 'Baseball', keywords: ['baseball', 'mlb', 'sports', 'recreation', 'tickets', 'ballpark'] },
	{ emoji: '🏒', name: 'Hockey', keywords: ['hockey', 'nhl', 'ice', 'sports', 'recreation', 'rink', 'tickets'] },
	{ emoji: '🎾', name: 'Tennis', keywords: ['tennis', 'sports', 'recreation', 'fitness', 'court', 'racket'] },
	{ emoji: '🏌️', name: 'Golf', keywords: ['golf', 'sports', 'recreation', 'course', 'club', 'driving range'] },
	{ emoji: '🏊', name: 'Swimming', keywords: ['swimming', 'pool', 'fitness', 'sports', 'recreation', 'aquatics', 'lap'] },
	{ emoji: '🧗', name: 'Climbing', keywords: ['climbing', 'rock climbing', 'fitness', 'sports', 'recreation', 'bouldering'] },
	{ emoji: '🏄', name: 'Surfing', keywords: ['surfing', 'surf', 'ocean', 'sports', 'recreation', 'beach', 'board'] },
	{ emoji: '🥊', name: 'Boxing', keywords: ['boxing', 'mma', 'martial arts', 'fitness', 'sports', 'gym', 'training'] },
	{ emoji: '🎿', name: 'Ski Gear', keywords: ['ski', 'snowboard', 'winter sports', 'gear', 'equipment', 'recreation'] },
	{ emoji: '🚴', name: 'Cycling', keywords: ['cycling', 'biking', 'peloton', 'fitness', 'sports', 'recreation', 'spin'] },
	{ emoji: '🤸', name: 'Gymnastics', keywords: ['gymnastics', 'fitness', 'sports', 'recreation', 'flexibility', 'pilates'] },

	// Events & Misc
	{ emoji: '🎉', name: 'Events', keywords: ['events', 'party', 'celebration', 'entertainment', 'tickets', 'concert', 'stubhub', 'ticketmaster'] },
	{ emoji: '📮', name: 'Mail', keywords: ['mail', 'postage', 'stamps', 'shipping', 'usps', 'fedex', 'ups'] },
	{ emoji: '🌱', name: 'Eco/Green', keywords: ['eco', 'green', 'sustainable', 'environment', 'solar', 'carbon', 'recycling'] },
	{ emoji: '☀️', name: 'Outdoor', keywords: ['outdoor', 'activities', 'recreation', 'park', 'hiking', 'camping', 'rei', 'patagonia'] },
	{ emoji: '🏔️', name: 'Hiking', keywords: ['hiking', 'mountain', 'trail', 'outdoor', 'recreation', 'camping', 'rei', 'national park'] },
	{ emoji: '🌊', name: 'Ocean', keywords: ['ocean', 'water sports', 'diving', 'snorkeling', 'beach', 'travel', 'recreation'] },
	{ emoji: '🎰', name: 'Gambling', keywords: ['gambling', 'casino', 'lottery', 'betting', 'poker', 'draftkings', 'fanduel'] },
	{ emoji: '🚬', name: 'Tobacco', keywords: ['tobacco', 'cigarettes', 'smoking', 'vaping', 'juul'] },
	{ emoji: '🍃', name: 'Cannabis', keywords: ['cannabis', 'dispensary', 'marijuana', 'weed', 'cbd'] },
	{ emoji: '🌙', name: 'Nightlife', keywords: ['nightlife', 'bar', 'club', 'nightclub', 'evening', 'entertainment'] },
	{ emoji: '🎄', name: 'Holiday', keywords: ['holiday', 'christmas', 'gifts', 'seasonal', 'celebration', 'decoration'] },
	{ emoji: '💌', name: 'Dating', keywords: ['dating', 'tinder', 'bumble', 'match', 'hinge', 'subscription', 'entertainment'] },
];
