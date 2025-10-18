'use client';

import { SiInstagram, SiTelegram, SiX, SiYoutube } from '@icons-pack/react-simple-icons';
import { Globe } from 'lucide-react';
import Image from 'next/image';

type Props = {
	url: string;
	size?: 3 | 4 | 5; // solo aceptamos 3,4,5
};

export function SocialLinkIcon({ url, size = 4 }: Props) {
	const lower = url.toLowerCase();

	let Icon: React.ComponentType<{ className?: string }> | string = Globe;

	if (lower.includes('youtube.com') || lower.includes('youtu.be')) {
		Icon = SiYoutube;
	} else if (lower.includes('instagram.com')) {
		Icon = SiInstagram;
	} else if (lower.includes('tiktok.com')) {
		Icon = '/providers/tiktok.webp';
		// Icon = SiTiktok;
	} else if (lower.includes('twitter.com') || lower.includes('x.com')) {
		Icon = SiX;
	} else if (lower.includes('t.me')) {
		Icon = SiTelegram;
	}

	const sizeClass = size === 3 ? 'h-3 w-3' : size === 5 ? 'h-5 w-5' : 'h-4 w-4';
	const px = size === 3 ? 12 : size === 5 ? 20 : 16; // para next/image

	return (
		<a href={url} target="_blank" rel="noreferrer" className="opacity-80 hover:opacity-100 flex items-center justify-center">
			{typeof Icon === 'string' ? (
				<Image src={Icon} alt="Social" width={px} height={px} className={sizeClass} />
			) : (
				<Icon className={sizeClass} />
			)}
		</a>
	);
}
