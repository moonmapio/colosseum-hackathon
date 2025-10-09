import acceptLanguage from 'accept-language';
import { NextRequest, NextResponse } from 'next/server';
import { cookieName, fallbackLng, headerName, languages } from './settings';

acceptLanguage.languages(languages);

export class I18n {
	private req: NextRequest;
	private url: URL;
	private pathname: string;
	private search: string;

	constructor(req: NextRequest) {
		this.req = req;
		this.url = req.nextUrl.clone();
		this.pathname = this.url.pathname;
		this.search = this.url.search;
	}

	private fromCookie(): string | undefined {
		if (!this.req.cookies.has(cookieName)) return;
		return acceptLanguage.get(this.req.cookies.get(cookieName)!.value) || undefined;
	}

	private fromHeader(): string | undefined {
		return acceptLanguage.get(this.req.headers.get('accept-language') || '') || undefined;
	}

	private detectLng(): string {
		return this.fromCookie() || this.fromHeader() || fallbackLng;
	}

	private langsInPath(): string | undefined {
		return languages.find((l) => this.pathname.startsWith(`/${l}`));
	}

	public ensureLocalePrefix(): NextResponse | null {
		const lngInPath = this.langsInPath();
		if (lngInPath || this.pathname.startsWith('/_next')) {
			return null;
		}
		const detected = this.detectLng();
		const target = new URL(`/${detected}${this.pathname}${this.search}`, this.req.url);
		return NextResponse.redirect(target);
	}

	public applyRefererCookie(res: NextResponse): NextResponse {
		const referer = this.req.headers.get('referer');
		if (!referer) return res;

		try {
			const refUrl = new URL(referer);
			const refLang = languages.find((l) => refUrl.pathname.startsWith(`/${l}`));
			if (refLang) {
				res.cookies.set(cookieName, refLang, {
					path: '/',
					sameSite: 'strict',
					maxAge: 60 * 60 * 24 * 30,
				});
			}
		} catch {
			// URL inválido: ignorar
		}
		return res;
	}

	public setLocaleHeader(res: NextResponse): NextResponse {
		const lngInPath = this.langsInPath() || this.detectLng();
		res.headers.set(headerName, lngInPath);
		return res;
	}

	public handle(): NextResponse | null {
		const redirect = this.ensureLocalePrefix();
		if (redirect) return redirect;

		return null;
	}

	public getCurrentLang(): string {
		return this.langsInPath() || this.detectLng();
	}
}
