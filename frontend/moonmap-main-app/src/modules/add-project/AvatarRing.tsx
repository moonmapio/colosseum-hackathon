'use client';

import { useAppStore } from '@modules/stores/main/app-hooks';
import { cn } from '@modules/utils';

type Props = { className?: string; ringState: '' | 'active' | 'ready' | 'failed' };

export function AvatarRing(props: Props) {
	const { className, ringState } = props;
	const state = useAppStore((state) => state.draftProjectLogo?.status);
	const isActive = state === 'pending' || state === 'uploaded' || state === 'processing';
	// const isReady = state === 'ready';
	const isFailed = state === 'failed';

	if (isActive || ringState === 'active') {
		return (
			<div className={cn('absolute inset-0 rounded-full pointer-events-none text-primary', className)} aria-hidden>
				<div
					className="absolute -inset-[2px] rounded-full animate-spin"
					style={{
						background: 'conic-gradient(from 0deg, currentColor 0 25%, transparent 25% 100%)',
						WebkitMask: 'radial-gradient(farthest-side, transparent calc(100% - 3px), black 0)',
						mask: 'radial-gradient(farthest-side, transparent calc(100% - 3px), black 0)',
					}}
				/>
			</div>
		);
	}

	// if (isReady || ringState === 'ready') {
	// 	return (
	// 		<div className={cn('absolute inset-0 rounded-full pointer-events-none text-emerald-100', className)} aria-hidden>
	// 			<div
	// 				className="absolute -inset-[2px] rounded-full"
	// 				style={{
	// 					background: 'conic-gradient(from 0deg, currentColor 0 100%, transparent 0% 100%)',
	// 					WebkitMask: 'radial-gradient(farthest-side, transparent calc(100% - 3px), black 0)',
	// 					mask: 'radial-gradient(farthest-side, transparent calc(100% - 3px), black 0)',
	// 				}}
	// 			/>

	// 			<div
	// 				className="absolute -inset-[2px] rounded-full text-emerald-200 animate-spin"
	// 				style={{
	// 					background: 'conic-gradient(from 0deg, currentColor 0 55%, transparent 25% 100%)',
	// 					WebkitMask: 'radial-gradient(farthest-side, transparent calc(100% - 3px), black 0)',
	// 					mask: 'radial-gradient(farthest-side, transparent calc(100% - 3px), black 0)',
	// 				}}
	// 			/>
	// 		</div>
	// 	);
	// }

	if (isFailed || ringState === 'failed') {
		return (
			<div className={cn('absolute inset-0 rounded-full pointer-events-none text-destructive', className)} aria-hidden>
				<div
					className="absolute -inset-[2px] rounded-full"
					style={{
						background: 'conic-gradient(from 0deg, currentColor 0 100%, transparent 25% 100%)',
						WebkitMask: 'radial-gradient(farthest-side, transparent calc(100% - 3px), black 0)',
						mask: 'radial-gradient(farthest-side, transparent calc(100% - 3px), black 0)',
					}}
				/>
			</div>
		);
	}

	return null;
}
