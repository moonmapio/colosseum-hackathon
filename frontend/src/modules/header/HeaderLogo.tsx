'use client';
import { useEphemeralAppStore } from '@modules/stores/main/ephemeral-hooks';
import Image from 'next/image';

export const HeaderLogo = () => {
	const isMobil = useEphemeralAppStore((state) => state.isMobile);
	if (!isMobil) return null;

	return <Image src={'/logo.svg'} alt={'MoonMap'} width={20} height={20} />;
};
