// LiveBadge.tsx
export function LiveBadge({ label = 'LIVE', className = '' }: { label?: string; className?: string }) {
	return (
		<span
			className={`inline-flex items-center gap-1 rounded-full bg-destructive/10 px-2 py-0.5 text-[10px] font-semibold text-primary ${className}`}
			aria-label="Transmisión activa"
		>
			<span className="relative flex h-2 w-2" aria-hidden>
				<span className="absolute inline-flex h-2 w-2 rounded-full bg-current opacity-75 animate-ping motion-reduce:animate-none" />
				<span className="relative inline-flex h-2 w-2 rounded-full bg-current" />
			</span>
			{label}
		</span>
	);
}
