'use client';

import { Avatar, AvatarImage } from '@modules/ui/avatar';
import { Skeleton } from '@modules/ui/skeleton';
import { cn } from '@modules/utils';
import { BadgeCheck } from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';

type Props = {
	src: string;
	alt: string;
	name?: string; // para iniciales en el fallback
	verified?: boolean;
	size?: number; // px
	className?: string;
	ring?: boolean;
};

// function initials(s?: string) {
// 	if (!s) return '•';
// 	const parts = s.trim().split(/\s+/);
// 	const a = parts[0]?.[0] ?? '';
// 	const b = parts[1]?.[0] ?? '';
// 	return (a + b || s[0]).toUpperCase();
// }

export function AvatarNext({ src, alt, name, verified, size = 48, className, ring = true }: Props) {
	const [loaded, setLoaded] = useState(false);
	const [err, setErr] = useState(false);

	useEffect(() => {
		setLoaded(false);
	}, [src]);

	const onLoad = useCallback(() => {
		setLoaded(true);
	}, []);

	const onError = useCallback(() => {
		setErr(true);
	}, []);

	const avatarClassName = useMemo(() => {
		return cn('h-full w-full object-cover rounded-full transition-opacity duration-300', loaded && !err ? 'opacity-100' : 'opacity-0');
	}, [err, loaded]);

	return (
		<div className={cn('relative inline-block', className)} style={{ width: size, height: size }}>
			{/* Skeleton mientras no cargue ni haya error */}
			{!loaded && !err && <Skeleton className="absolute inset-0 rounded-full animate-pulse" />}

			<Avatar className={cn('h-full w-full rounded-full', ring && 'ring-1 ring-border')}>
				<AvatarImage src={src} alt={alt} onLoad={onLoad} onError={onError} className={avatarClassName} />
				{/* <AvatarFallback className="rounded-full bg-muted text-xs font-medium object-cover">{initials(name || alt)}</AvatarFallback> */}
			</Avatar>

			{verified && (
				<BadgeCheck
					fill="currentColor"
					stroke="white"
					strokeWidth={2}
					className="absolute -bottom-1 -right-1 size-6 text-primary drop-shadow"
				/>
			)}
		</div>
	);
}
