'use client';

import i18next from './i18next';
import { useParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { UseTranslationOptions } from 'react-i18next';

const runsOnServerSide = typeof window === 'undefined';

export function useT(ns: string, options?: UseTranslationOptions<undefined>) {
	const params = useParams();
	const lng = params.language;

	const [activeLng, setActiveLng] = useState(i18next.resolvedLanguage);
	if (typeof lng !== 'string') throw new Error('useT is only available inside /app/[lng]');

	const shouldChange = runsOnServerSide && i18next.resolvedLanguage !== lng;

	if (shouldChange) {
		i18next.changeLanguage(lng);
	}

	useEffect(() => {
		if (shouldChange) return;
		if (activeLng === i18next.resolvedLanguage) return;
		setActiveLng(i18next.resolvedLanguage);
	}, [activeLng, shouldChange]);

	useEffect(() => {
		if (shouldChange) return;
		if (!lng || i18next.resolvedLanguage === lng) return;
		i18next.changeLanguage(lng);
	}, [lng, shouldChange]);

	return useTranslation(ns, options);
}
