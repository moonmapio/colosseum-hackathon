// VideoPulseIcon.tsx
import { cn } from '@modules/utils';
import { Video } from 'lucide-react';

export function VideoPulseIcon({
	active = true,
	className = 'text-primary',
	size = 20,
}: {
	active?: boolean;
	className?: string;
	size?: number;
}) {
	return (
		<span
			className={cn(`relative inline-flex items-center justify-center ${className} `, !active ? '!text-muted-foreground' : '')}
			aria-label="ON AIR"
		>
			{active && <span className="absolute h-6 w-6 rounded-full bg-current/20 animate-ping motion-reduce:animate-none" aria-hidden />}
			<Video className="relative z-10" style={{ width: size, height: size }} aria-hidden />
		</span>
	);
}
