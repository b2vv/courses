# Research: кодеки, контейнери й ladder у великих стримінгах — і де сидить AV1

Дослідницька нотатка, 2026-09-06. Тільки первинні джерела (інженерні блоги, офіційні
developer docs, опубліковані специфікації, white papers); вторинні згадки позначені явно.
Нотатка не змінює рішень Phase 5 (`docs/media/spec.md` Q3, `docs/architecture/media.md`,
`docs/adr/005-hls-nginx.md`) — вона їх перевіряє проти того, що робить індустрія.

Контекст платформи: ~нуль користувачів, локальний диск, без CDN, один media-worker з
FFmpeg і single-concurrency, плеєр з'явиться лише у Phase 6.

---

## 1. Вердикт на один екран

**Що кодувати сьогодні.** Один кодек, один контейнер, три сходинки:

| Рішення | Значення | Чому (з доказами нижче) |
|---|---|---|
| Video codec | **H.264/AVC High Profile, Level ≤ 4.1**, 8-bit 4:2:0 | Apple: «for maximum compatibility some H.264 variants SHOULD be ≤ High 4.1» (1.3a), «SHOULD use High Profile» (1.4), «for backward compatibility some video content SHOULD be encoded with H.264» (1.12) [S7]. Mux — комерційний енкодер для таких як ми — віддає глядачам лише H.264, навіть 4K [S27][S30]. Bitmovin VDR: H.264 у production у 84–86 % респондентів, «far and away the most popular» [S23]. |
| Audio codec | **AAC-LC**, stereo | Apple 2.3: «Stereo audio in AAC, HE-AAC v1, or HE-AAC v2 format MUST be provided» [S7]; YouTube ingest — AAC-LC [S12]; Mux «only delivers AAC audio» [S27]. |
| Container / segments | **HLS з fMP4 (CMAF-style, `.m4s`)**, video і audio окремими треками | Apple 1.2 дозволяє обидва варіанти для H.264, але 1.5/1.39 — «container format for HEVC/AV1 video MUST be fMP4» [S7]; тобто fMP4 — єдиний контейнер, який не доведеться міняти, якщо колись додаємо другий кодек. Apple 9.5: «SHOULD deliver video and audio as separate streams». VDR: HLS у production 80 % (VOD), DASH 69 % [S23] — обидва живуть на тих самих fMP4 байтах. |
| Segmenting | target duration **6 s**, IDR **кожні 2 s**, `EXT-X-INDEPENDENT-SEGMENTS`, `AVERAGE-BANDWIDTH` + `CODECS` + `RESOLUTION` + `FRAME-RATE` у master | Apple 7.5–7.7, 1.13, 9.1–9.2, 9.11, 9.14, 9.15 [S7]. |
| Ladder | **360p / 720p / 1080p** без upscaling; орієнтири бітрейту — Apple H.264 таблиця (640×360 ≈ 365 kbit/s; 1280×720 ≈ 3000–4500; 1920×1080 ≈ 6000–7800) з CRF-режимом x264 замість жорсткого CBR | Apple 1.25 (таблиця «one possible set»), 9.9 «MUST provide multiple bit rates», 9.10 «for HD content SHOULD be at least two variants at the highest resolution» [S7]. Netflix показав, що фіксована ladder марнує біти на простому контенті й не вистачає на складному (per-title, 2015) [S4] — для нас CRF дає «безкоштовний» per-title ефект без convex-hull пошуку. |
| HEVC | **Ні** | Виграш лише на Apple-пристроях (fMP4-only, 1.5) [S7]; Disney+ у серпні 2026 втратив 4K/HDR в 11 країнах ЄС через injunction UPC по HEVC-патентах і повертає 4K через VP9 [S32][S33]; Mux досі не віддає HEVC на вихід [S27][S28]. |
| AV1 | **Не зараз; тільки як додатковий ряд, ніколи як заміна** | Див. тригер нижче. Всі гіганти тримають H.264 як базу: Netflix — «primarily relied on H.264/AVC», AVC досі #1 [S1][S6]; Apple 1.12/1.24; Bitmovin: «anyone encoding with AV1 will still need to provide backup H.264/AVC streams» [S25]. |

**Тригер додати AV1** (усі три умови одночасно, і лише як другий ряд поруч із H.264):

1. **Є що економити.** З'явився реальний рахунок за egress (CDN або платний трафік) або
   є виміряні сесії, що впираються у bandwidth. Поки байти йдуть з локального диска через
   nginx без CDN, ~30 % економії бітрейту [S1][S11][S19] економлять нуль гривень.
2. **Є телеметрія пристроїв.** Phase 6 плеєр звітує `MediaCapabilities`/CODECS-вибір, і
   частка сесій з AV1-декодером стабільно ≥ 50 %. Без цього ми не знаємо, чи не кодуємо
   AV1 для нікого. Netflix і Meta включали AV1 саме по пристроях і саме по телеметрії
   [S2][S19].
3. **Є compute.** Worker витримує другий прохід: SVT-AV1 preset 8–10 (документація
   SVT-AV1: presets 7–13 «for fast and real-time encoding») [S36] на тому ж джерелі, без
   зриву SLA на publish. Netflix прямо каже, що AV1 «takes more CPU hours» і потребував
   окремої стратегії rollout по популярності тайтлів [S2]; Meta вмикала AV1 лише коли
   «projected watch time exceeds the threshold» [S19].

Коли тригер спрацював — додаємо AV1 у той самий fMP4 master playlist як окремий набір
варіантів (`CODECS="av01.0.…"`), Apple 1.23: «each stream SHOULD provide all anticipated
bandwidths. Clients SHOULD NOT be required to switch codecs» [S7]. Контейнер міняти не
доведеться — саме тому fMP4 обирається сьогодні.

---

## 2. Факти по сервісах

Кожен рядок несе URL джерела. «Additive» = AV1 додається поруч зі старими кодеками;
«replacing» = старий кодек прибирають.

### 2.1 Netflix

| Факт | Значення | Джерело |
|---|---|---|
| Базовий кодек | «Since entering the streaming business in 2007, Netflix has primarily relied on H.264/AVC». Станом на 03-2025: «AV1-SDR is already the second most streamed codec at Netflix, behind H.264/AVC» | [S1](https://netflixtechblog.com/av1-now-powering-30-of-netflix-streaming-02f592242d80), [S6](https://netflixtechblog.com/hdr10-now-streaming-on-netflix-c9ab1f4bd72b) |
| Ladder у production | AVC (Main для старих девайсів, High «AVCHi-Opt» з 2018), VP9 (mobile, з 2016), HEVC (HDR10 / Dolby Vision), AV1 (SDR з 2020/2021, HDR10+ з 03-2025) | [S5](https://netflixtechblog.com/optimized-shot-based-encodes-now-streaming-4b9464204830), [S6], [S1] |
| Per-title (2015) | Фіксована ladder 2010 року мала верх 1080p @ 5800 kbps, 480p @ 1750 kbps; резолюції 1920×1080, 1280×720, 720×480, 512×384, 384×288, 320×240, крок бітрейту ~5 %. Per-title: *Orange is the New Black* 1080p 4640 kbps замість 5800 (−20 %); *BoJack* — 1080p уже з 1540 kbps замість 480p @ 1750 | [S4](https://netflixtechblog.com/per-title-encode-optimization-7e99442b62a2) |
| Per-shot / Dynamic Optimizer (2018) | Шот ≈ 4 s; VMAF як objective; при VMAF 80 AVCHi-Opt потребує «less than half of the bits» AVCMain, VP9-Opt — «less than one third»; на ~250 kbps VMAF 58 (AVCMain) → 73 (AVCHi-Opt) → 79 (VP9-Opt). Кількість encode units зросла «more than two orders of magnitude» | [S5] |
| AV1 rollout — коли/де | 02-2020 Android app, «Save Data», dav1d software decoder, «20 % improved compression efficiency over our VP9 encodes» → 11-2021 TV (hardware decoders, PS4 Pro) → 2022 web browsers → 2023 Apple A17 Pro / M3 → 03-2025 AV1-HDR10+ | [S3](https://netflixtechblog.com/netflix-now-streaming-av1-on-android-d5264a515202), [S2](https://netflixtechblog.com/bringing-av1-streaming-to-netflix-members-tvs-b7fc88e42320), [S1], [S6] |
| AV1 recipe | Завжди у максимальній source-резолюції та fps; **усі AV1-потоки 10-bit**; shot-level dynamic optimization; власні certification streams для декодерів | [S2] |
| Опубліковані числа (12-2025) | AV1 ≈ 30 % усього перегляду, «second most-used codec», VMAF +4.3 vs AVC, +0.9 vs HEVC; «one-third less bandwidth than both AVC and HEVC»; «45 % fewer buffering interruptions»; dav1d дає ≈ 40 % browser-playback; 88 % large-screen пристроїв на сертифікації 2021–2025 підтримують AV1, «since 2023, almost all»; FGS: 2804 kbps замість 8274 (−66 %) | [S1] |
| Числа A/B 2021 (TV) | до +10 VMAF у найгірших сесіях, 4K-тривалість +5 %, quality drops −38 % на деяких TV, play delay −2 % | [S2] |
| Encode cost | «AV1 encoding takes more CPU hours»; rollout по популярності тайтлів; окрема CPU-scheduling стратегія. Кратність не публікується | [S2] |
| Additive чи replacing | Additive: AVC залишається базою; VP9/HEVC теж живі; на ingest Netflix і далі приймає лише Dolby Vision masters | [S1], [S6] |
| Container | ⚠️ неперевірено — у прочитаних постах формат пакування (ISOBMFF/CMAF vs TS) не названий |  |
| Драйвер вибору | Власний CDN Open Connect + «more than 300 million members» → bandwidth і QoE; сертифікація пристроїв з 2019 | [S1] |

### 2.2 Apple — HLS Authoring Specification, Apple TV+ і AV1-залізо

| Факт | Значення | Джерело |
|---|---|---|
| Дозволені кодеки | 1.1 «All video MUST be encoded using H.264/AVC, HEVC/H.265, Dolby Vision, or AV1» | [S7](https://developer.apple.com/documentation/http-live-streaming/hls-authoring-specification-for-apple-devices) |
| H.264 | 1.2 контейнер «fMP4 files or MPEG transport streams»; 1.3a «some variants SHOULD be ≤ High Profile, Level 4.1»; 1.3b «MUST be ≤ High 5.2»; 1.4 «SHOULD use High Profile»; 1.12 «for backward compatibility, some video content SHOULD be encoded with H.264»; 1.10 `avc1` (parameter sets у sample description), не `avc3` | [S7] |
| HEVC / AV1 | 1.5 «container format for HEVC video MUST be fMP4»; 1.6a ≤ Main 10 L4.0; 1.37 «Level for AV1 MUST be ≤ 6.2»; 1.39 «container format for AV1 video MUST be fMP4» | [S7] |
| Multi-codec | 1.23 «If multiple video streams are provided (H.264, HEVC, HDR), each stream SHOULD provide all anticipated bandwidths. Clients SHOULD NOT be required to switch codecs»; 1.24 «SDR streams MUST be provided» | [S7] |
| Audio | 2.3 «Stereo audio in AAC, HE-AAC v1, or HE-AAC v2 format MUST be provided»; 2.4 «SHOULD NOT use HE-AAC if audio bit rate > 64 kbit/s»; 2.25 xHE-AAC/FLAC/ALAC — fMP4 only | [S7] |
| Segments | 1.13 IDR «every two seconds»; 7.4 «Video segments MUST start with an IDR»; 7.5 «Target durations SHOULD be 6 seconds»; 7.7 сегмент не довший за target + 0.5 s; 9.11/9.12 `EXT-X-INDEPENDENT-SEGMENTS` | [S7] |
| Playlist | 9.1 `CODECS` MUST; 9.2 `RESOLUTION` MUST; 9.5 «SHOULD deliver video and audio as separate streams»; 9.9 «MUST provide multiple bit rates»; 9.10 ≥ 2 варіанти на найвищій резолюції для HD; 9.13/9.14 `BANDWIDTH` = peak, `AVERAGE-BANDWIDTH` MUST; 9.15 `FRAME-RATE` MUST; 1.26/1.27 для VOD середній/піковий бітрейт у межах 10 % від оголошених; 1.32 default-варіант ≈ 2000 kbit/s | [S7] |
| Ladder (H.264, «one possible set», 16:9) | 416×234 @ 145 · 640×360 @ 365 · 768×432 @ 730 · 768×432 @ 1100 · 960×540 @ 2000 · 1280×720 @ 3000 · 1280×720 @ 4500 · 1920×1080 @ 6000 · 1920×1080 @ 7800 kbit/s | [S7] |
| Ladder (HEVC 30 fps) | 640×360 @ 145 · 768×432 @ 300 · 960×540 @ 600/900/1600 · 1280×720 @ 2400/3400 · 1920×1080 @ 4500/5800 · 2560×1440 @ 8100 · 3840×2160 @ 11600/16800 kbit/s; «24 fps HEVC content should use bit rates reduced by about 20 %» | [S7] |
| MIME | `application/vnd.apple.mpegurl` (.m3u8), `video/mp2t` (.ts), `video/mp4` (init), `video/iso.segment` (.m4s) | [S7] |
| Trick play | 6.1 I-frame playlists MUST; 6.14/6.16 H.264 та SDR trick play для сумісності | [S7] |
| Cellular amendment | 9.21 варіант ≤ 192 kbit/s MUST для cellular; 9.22 «a peak 192 kbit/s H.264 variant packaged in a transport stream SHOULD be provided» — єдине місце, де TS усе ще рекомендований | [S7] |
| AV1-залізо | A17 Pro (iPhone 15 Pro, 2023-09-12): «includes a dedicated AV1 decoder»; M3 (2023-10-30): «media engine now includes support for AV1 decode» | [S8](https://www.apple.com/newsroom/2023/09/apple-unveils-iphone-15-pro-and-iphone-15-pro-max/), [S9](https://www.apple.com/newsroom/2023/10/apple-unveils-m3-m3-pro-and-m3-max-the-most-advanced-chips-for-a-personal-computer/) |
| Safari | Safari 17.0 (2023-09-18): «adds support for AV1 video on devices with hardware decoding support, like iPhone 15 Pro» — тобто **software AV1 у Safari немає**; старіші iPhone/Mac AV1 не грають | [S10](https://webkit.org/blog/14445/webkit-features-in-safari-17-0/) |
| Apple TV+ production ladder | ⚠️ неперевірено — Apple не публікує кодеки/ladder Apple TV+; вище — лише вимоги до сторонніх HLS-стрімів |  |

### 2.3 YouTube / Google

| Факт | Значення | Джерело |
|---|---|---|
| Ingest (upload) | Рекомендований upload: H.264 High Profile, closed GOP «of half the frame rate», CABAC, 4:2:0, MP4 з moov на початку; audio AAC-LC/Opus 48 kHz; SDR 1080p 8 Mbps (12 при HFR), 720p 5 (7.5), 480p 2.5 (4), 360p 1 (1.5). AV1 не згадується як upload-формат; «To view new 4K uploads in 4K, use a browser or device that supports VP9» | [S12](https://support.google.com/youtube/answer/1722171?hl=en) |
| Ingest (live) | Кодеки live-енкодера: «H.264, H.265 (HEVC), AV1»; keyframe «Recommended 2 seconds. Do not exceed 4 seconds»; 1080p60 12 Mbps, 1080p30 10, 720p60 6, 240p–720p30 4 Mbps. Таблиця протоколів: RTMP/RTMPS — H.264; HLS — H.264, H.265; DASH — H.264, VP9 | [S13](https://support.google.com/youtube/answer/2853702?hl=en), [S14](https://developers.google.com/youtube/v3/live/guides/ingestion-protocol-comparison) |
| AV1 rollout (за спільним white paper Meta/Vodafone/YouTube, 09-2025) | «YouTube launched AV1 for VOD playbacks on desktop web browsers in 2018 using software decode … then launched AV1 for televisions and mobile phones as AV1 hardware support became available, starting in 2023. In 2024, YouTube began deploying software AV1 decode for Android and iOS devices…»; «In 2025, YouTube launched AV1 support for popular live streams»; «Today, over 75 % of YouTube's catalog (weighted by watch time) is available in AV1» | [S11](https://engineering.fb.com/wp-content/uploads/2025/09/Meta-AV1-White-Paper-FINAL.pdf) |
| Власне залізо для транскодування | «YouTube's custom hardware video encoder reduces by over 95 % the amount of electrical power used to convert creators' uploads» | [S11] |
| Android як екосистема | Android 14 CDD: handheld/TV **MUST** декодувати AV1 (5.3/H-0-6, 5.3.2/T-0-7) і кодувати AV1 (5.2/H-0-3); Performance Class U: «MUST support AV1 hardware decoder Main 10, Level 4.1 and film grain» | [S17](https://source.android.com/docs/compatibility/14/android-14-cdd) |
| dav1d на Android | Android 15 Beta 2 (05-2024): dav1d «up to 3x more performant than the legacy AV1 software decoder», backported до Android 11 через Google Play system updates; на той момент opt-in `c2.android.av1-dav1d.decoder`, «will be made the default … in a subsequent update» | [S16](https://android-developers.googleblog.com/2024/05/the-second-beta-of-android-15.html) |
| Chrome / Firefox | Chrome 70 (2018-09): «AV1 decoder is now supported in Chrome Desktop x86-64»; Firefox 67 (2019-05-21): «upgrading to the newer, higher performance, AV1 decoder known as 'dav1d'» | [S15](https://developer.chrome.com/blog/media-updates-in-chrome-70), [S34](https://www.firefox.com/en-US/firefox/67.0/releasenotes/) |
| Стейтмент AOMedia | «YouTube has already streamed millions of hours of HDR video using AV1» | [S18](https://aomedia.org/av1-adoption-showcase/google-story/) |
| Ladder / policy у продакшні | ⚠️ неперевірено — YouTube не публікує ladder і правило «коли AV1 замість VP9/H.264»; налаштування «Prefer AV1 for SD / Always prefer AV1» видно лише в залогіненому UI, офіційної довідкової сторінки не знайдено. Заяви «YouTube > 75 % playback на AV1» у пресі — вторинні і не збігаються з формулюванням white paper (75 % **каталогу за watch time доступно** в AV1) |  |
| Container | ⚠️ неперевірено в первинних джерелах (YouTube DASH/ISOBMFF-подібна доставка не описана офіційно) |  |
| Additive чи replacing | Additive: H.264 лишається ingest-форматом і базою відтворення; VP9 і AV1 — додаткові транскоди | [S12], [S11] |

### 2.4 Meta (Facebook / Instagram Reels) — бонус, бо публікує найконкретніші числа

| Факт | Значення | Джерело |
|---|---|---|
| Кодеки | «three encoding families with AVC, VP9, and AV1» | [S19](https://engineering.fb.com/2023/02/21/video-engineering/av1-codec-facebook-instagram-reels/) |
| Encoder / decoder | SVT-AV1 (13 presets); dav1d в плеєрі iOS та Android; dav1d тримає 720p30 на більшості Android, 1080p30 на mid/high | [S19] |
| Числа | «about 30 % better coding efficiency than VP9 and HEVC»; приклад: −65 % vs AVC, −48 % vs VP9; на iPhone Reels середній бітрейт −12 %, FB-MOS +0.6; AV1 «> 70 % of global video watch on the Meta Family of Apps» | [S19], [S11] |
| Кому кодують AV1 | Двостадійно: «When projected watch time of the video exceeds the threshold, second stage encoding is triggered» | [S19] |
| Encode cost як енергія | «SVT-AV1 has the convex hull over 2.5 orders of magnitude, delivering ~40 % fewer bits at the same power cost as x264 'veryfast' … ~30 % fewer bits than x265 at the same energy level»; high-tier телефони з HW AV1 — ~25 % нових пристроїв, low/mid — ~75 % | [S11] |

### 2.5 Twitch

| Факт | Значення | Джерело |
|---|---|---|
| Base | H.264; «Enhanced Broadcasting» (beta 2024-01-08) переносить multi-rendition encode на GPU стрімера; HEVC/AV1 — «experimenting», AV1 «exclusively available with NVIDIA GeForce RTX 40 Series GPUs» | [S20](https://blog.twitch.tv/en/2024/01/08/introducing-the-enhanced-broadcasting-beta) |
| 2026-06-17 | 2K (1440p) для Partners/Affiliates **на HEVC** «up to 9 Mbps for 1440p and 7.5 Mbps for 1080p»; server-side transcode для 2K через Enhanced Broadcasting | [S21](https://blog.twitch.tv/en/2026/06/17/introducing-dual-format-and-2k-streaming-on-twitch/) |
| AV1 для глядачів | ⚠️ неперевірено — у блозі 06-2026 AV1 не згадано; твердження «AV1 і 4K лише для beta community» знайдено тільки у вторинних джерелах; help-сторінка Twitch не віддалась |  |
| Container | ⚠️ неперевірено в первинних джерелах |  |
| Драйвер | Live: encode-cost перенесений на клієнта (GPU NVENC), HEVC для 1440p через ліміт бітрейту, а не через AV1 | [S20], [S21] |

### 2.6 Vimeo

| Факт | Значення | Джерело |
|---|---|---|
| AV1 | 2019-06-13: AV1 лише для Staff Picks, «on any supported platform (including the most recent Chrome and Firefox)»; encoder rav1e (співпраця з Mozilla); H.264 лишається default | [S22](https://vimeo.com/blog/post/av1-new-standard-codecs) |
| Числа | Не опубліковані; Medium-пост «Behind the scenes of AV1 at Vimeo» повертає 410 («post suspended») | ⚠️ неперевірено |

### 2.7 Bitmovin і Mux (вендори з даними)

| Факт | Значення | Джерело |
|---|---|---|
| Bitmovin VDR 7 (2023/24) | «H.264/AVC is still far and away the most popular video codec»; у production H.264 84 % (VOD) / 86 % (live), HEVC 48 % / 46 %; «Adoption of newer generation codecs is still disappointingly low, but AV1 seems to have the most optimism … likely driven by Apple». Формати VOD: HLS 80 %, DASH 69 % («from 51 % last year»); live: «CMAF DASH/HLS surpassing RTMP passthrough». Точні AV1-відсотки з PDF витягуються неоднозначно (≈14–18 % production, ≈27–34 % planning) — ⚠️ звірити з оригінальним графіком | [S23](https://bitmovin.com/downloads/bitmovin-7th-video-developer-report-2023-2024.pdf) |
| Bitmovin VDR 8/9 | 9-й звіт (2025-09-11) — codec-цифри лише у повному PDF за формою; у пресі: H.264 79 %, HEVC 49 %, AV1 13 % production / 32 % planning, 167 респондентів з 34 країн — ⚠️ вторинне переказування | [S37](https://bitmovin.com/press-room/bitmovin-com-video-developer-report-2025/) |
| Bitmovin про AV1 | «~50 % less data» vs H.264, «~30 %» vs HEVC/VP9; «anyone encoding with AV1 will still need to provide backup H.264/AVC streams»; «AV1 will need to be part of a multi-codec approach for the foreseeable future»; libaom на ноутбуку 0.032 fps vs 7 fps у хмарі vs 36 fps у кластері (2024) | [S25](https://bitmovin.com/blog/av1-encoding-guide/), [S24](https://bitmovin.com/blog/av1-playback-support/), [S26](https://bitmovin.com/blog/bitmovin-improves-av1-video-encoding/) |
| Mux — вхід | Standard input: «H.264 or HEVC video codecs», інше «standardized automatically to H.264»; 1080p ≤ 8 Mbps, GOP ≤ 20 s (10 для HEVC), closed GOP, 8-bit 4:2:0 (10-bit лише HEVC), AAC. HEVC став standard input 2025-10-15 | [S27](https://www.mux.com/docs/guides/minimize-processing-time), [S28](https://www.mux.com/docs/changelog/hevc-standard-input-general-availability) |
| Mux — вихід | «Mux uses high-bitrate H.264 for delivering 4K content, which is supported on a wide range of devices»; «Mux only delivers AAC audio»; ladder за тарифом — «reduced / standard / extended», деталі не публікуються; HEVC/AV1 як output — не задокументовано | [S30](https://www.mux.com/docs/guides/video.txt), [S27], [S29](https://www.mux.com/docs/guides/use-video-quality-levels) |

### 2.8 Disney+ і Amazon Prime Video — те, що вдалося підтвердити

| Факт | Значення | Джерело |
|---|---|---|
| Disney+ базові кодеки | ⚠️ неперевірено первинно (Disney не публікує ladder; help-центр рендериться JS). Але подія серпня 2026 фіксує факт: 4K/HDR йшли на **HEVC**; після injunction UPC (InterDigital, Düsseldorf, 2026-07-23) Disney заявив: «We have implemented an alternative technology to enable 4K UHD and HDR … 4K UHD remains unavailable on certain devices»; 4K повернули на Fire TV, Android/Google TV, LG 2020+, Sky — тобто на VP9-сумісних пристроях; Apple TV 4K/Samsung/консолі — ні | [S32](https://flatpanelshd.com/news.php?id=1785834485&subaction=showfull), [S33](https://www.heise.de/en/news/Patent-dispute-Disney-restores-4K-in-Germany-11395286.html) (журналістика з прямою цитатою Disney) |
| Prime Video кодеки | ⚠️ неперевірено — Amazon не публікує кодеки/ladder Prime Video; преса каже «resolved a similar dispute by reaching a licensing agreement with InterDigital» [S32] |  |
| Amazon Fire TV (device side, первинно) | AV1 HW decode «up to 3840×2160p @ 60 fps, 100 Mbps» на Fire TV Stick 4K Max (2021, 2023), 4K (2023), 4K Plus/Select (2025), Stick HD (2026); Sticks 2014–2020 та Fire TV box 2014–2017 — AV1 не вказано; H.264/HEVC на всіх з 2015 | [S31](https://developer.amazon.com/docs/device-specs/device-specifications-fire-tv-streaming-media-player.html) |

### 2.9 Інструменти, які ми реально запускаємо

| Факт | Значення | Джерело |
|---|---|---|
| FFmpeg + AV1 | Три енкодери: `libaom-av1` (reference, «based on libvpx»), `libsvtav1` (Intel + Netflix, з 2020 — базa AOMedia), `librav1e`; libaom `-cpu-used` 0–8 («Lower values mean slower encoding»), realtime тільки з `-usage realtime`; SVT-AV1: CRF за замовчуванням, `-svtav1-params` з FFmpeg 5.1 | [S35](https://trac.ffmpeg.org/wiki/Encode/AV1) |
| SVT-AV1 presets | «presets 1–3 represent extremely high efficiency … Presets 4–6 are commonly used by home enthusiasts … Presets between 7 and 13 are used for fast and real-time encoding. One should use the lowest preset that is tolerable» | [S36](https://gitlab.com/AOMediaCodec/SVT-AV1/-/blob/master/Docs/CommonQuestions.md) |
| Кратність encode-time AV1 vs x264 | ⚠️ неперевірено — первинного числа «у N разів повільніше» не знайдено; єдине первинне порівняння — енергетичне (Meta/YouTube: ~40 % менше біт за ту саму енергію, що x264 veryfast) [S11]. Для нашого worker'а це треба міряти самим |

---

## 3. Що роблять гіганти і що НЕ переноситься на нас

| Практика гігантів | Чому в них це працює | Чому в нас — ні (поки) |
|---|---|---|
| **Три-чотири кодеки паралельно** (Netflix AVC+VP9+HEVC+AV1 [S1][S5]; Meta AVC+VP9+AV1 [S19]; YouTube H.264+VP9+AV1 [S12][S11]) | Кожен зайвий encode окупається мільярдами переглядів; егрес — головна стаття витрат («more than 300 million members» [S1]; «over 2.5 orders of magnitude» енергетичних порівнянь [S11]) | Наш егрес = локальний диск → nginx → кілька людей. Другий кодек подвоює час worker'а і місце на диску, економить нуль. |
| **Per-title / per-shot ladder з convex hull і VMAF** [S4][S5] | Потребує десятків пробних encode на тайтл і pipeline, що «increased the number of chunks by more than two orders of magnitude» [S5] | Single-concurrency worker. Правильний еквівалент для нас — CRF у x264: біти самі течуть до складних сцен, без пошуку. |
| **Власний CDN і device certification** (Open Connect; сертифікаційні AV1-потоки з 2019 [S1][S2]) | Netflix знає точну частку AV1-здатних пристроїв (88 % на сертифікації) і має A/B-платформу для codec-вибору | Ми не знаємо, що в нас за пристрої, поки Phase 6 не привезе телеметрію плеєра. Без неї «AV1 для кого?» — питання без відповіді. |
| **Rollout по популярності тайтлів** ([S2]; Meta «projected watch time exceeds the threshold» [S19]) | Compute витрачається лише там, де є перегляди | В нас нема перегляду, щоб рахувати threshold. Якщо колись — це готова стратегія: AV1 лише для лекцій з N+ переглядів. |
| **Власне залізо для транскодування** (YouTube: −95 % енергії на upload-конвертацію [S11]) | Масштаб дозволяє ASIC | Ми маємо один FFmpeg. Тому дорогі presets (SVT-AV1 1–3, libaom cpu-used 0–1) для нас нереалістичні; якщо AV1 — то preset 8+. |
| **10-bit AV1 всюди** [S2] | Джерела Netflix — 10-bit+ masters; 10-bit ловить banding | Наші джерела — screen-recording/камера 8-bit; 10-bit H.264 (High 10) ламає сумісність із hardware-декодерами. Лишаємось 8-bit 4:2:0. |
| **HEVC для HDR/Dolby Vision** [S6], Twitch HEVC для 1440p [S21] | HDR-каталог, ліцензії, юристи | HDR у нас нема; HEVC-патентний ризик (Disney+ UPC 2026 [S32][S33]) — не наша битва. |
| **Enhanced Broadcasting** — encode на GPU стрімера [S20][S21] | Live-платформа перекладає compute на клієнта | Ми VOD; upload — оригінал, encode — на сервері. Не застосовно. |
| **Software AV1 decode (dav1d) як fallback у власному плеєрі** [S1][S19][S16] | Власні нативні апки; можна вшити dav1d | Ми — браузер + hls.js. Safari має AV1 лише з hardware [S10]; ми не контролюємо декодер. |

Що **переноситься** дослівно: Apple HLS Authoring Spec як чекліст на packaging [S7];
CRF замість CBR; H.264 High ≤ 4.1 як база; fMP4, щоб не перепаковувати при другому кодеку;
«multi-codec = additive, never replacing» [S7 1.12/1.24][S25]; increment по пристроях і
телеметрії, а не по календарю.

---

## 4. Не вдалося перевірити (⚠️ неперевірено)

- Формат пакування (CMAF/ISOBMFF vs TS) у Netflix, YouTube, Twitch — жодне первинне
  джерело не описує.
- Production ladder і codec-policy YouTube («коли AV1 замість VP9»); офіційна довідкова
  сторінка для «Prefer AV1» налаштувань не знайдена.
- Кодеки й ladder Apple TV+ (Apple не публікує), Disney+ і Prime Video (жодного
  інженерного блогу; факти лише через журналістику з цитатами Disney [S32][S33]).
- Статус AV1 для глядачів Twitch у 2026 (блог мовчить, help-сторінка не рендериться).
- Числа Vimeo по AV1 (Medium-пост видалений, 410).
- Точні AV1-відсотки у Bitmovin VDR 7 (PDF-екстракція графіка неоднозначна) і всі цифри
  VDR 8/9 (тільки вторинний переказ).
- ~~Кратність encode-time AV1 vs x264 у FFmpeg — первинної цифри немає; міряти на своєму
  worker'і.~~ **Закрито 2026-09-20 власним виміром — див. §6.**
- Частка браузерів/пристроїв з AV1-декодером серед *наших* користувачів — даних нема
  до Phase 6.

---

## 6. Власний вимір, 2026-09-20 (AV1 vs наш чинний H.264)

Зроблено на машині розробника (arm64, macOS 26.5.2; ffmpeg 7.1.1, SVT-AV1 3.0.2,
libvmaf 3.0.0), **не на воркері** — воркер ships ffmpeg 5.1.9, тож числа часу
переносяться на нього лише як порядок величини, не як факт. Бік H.264 дзеркалить
`transcoder.go` дослівно: `libx264`, `profile high`, `yuv420p`, `crf 21`, maxrate/bufsize
та level рангу, `preset veryfast`, форсований keyframe кожні 2 с.

### 6.1 Поправка до рамки питання

Вимір спершу ставився як «замінити H.264 на AV1» — це **суперечить §24.14**, яка каже
«тільки як додаткова сходинка поряд, ніколи замість». Наслідок, який змінює висновок:
додаткова сходинка **збільшує** місце на диску (зберігаємо обидві), а економить **egress**
для тих клієнтів, які вміють AV1. Тому числа нижче — це вигода на сесію в умові (1)
§24.14, а не економія диска.

### 6.2 Кратність encode-time (закриває відкритий пункт §4, умова (3) §24.14)

Зміряно **на обох середовищах**, однакове джерело (720p25, 30 с), однакові налаштування,
строго послідовно — перша спроба ганяла їх паралельно, і базова лінія H.264 зсунулась
утричі, тобто зіпсувала саме те, що міряється.

| | H.264 veryfast | AV1 preset 8 | preset 6 | preset 4 |
|---|---|---|---|---|
| Хост (ffmpeg 7.1.1, **SVT-AV1 3.0.2**) | 1.60 с | 2.88 с — **1.8x** | 6.27 с — 3.9x | 28.37 с — 17.8x |
| **Воркер** (ffmpeg 5.1.9, **SVT-AV1 1.4.1**) | 3.10 с | 72.16 с — **23.3x** | 152.66 с — 49.2x | 565.05 с — **182x** |

**Умова (3) на поточному образі воркера НЕ виконується.** Кратність там у 13 разів гірша
за хостову. Це не накладні контейнера: x264 у воркері повільніший лише вдвічі, тож із
25-кратного розриву на AV1 близько 12x дає **версія кодера** — Debian bookworm постачає
SVT-AV1 1.4.1, а швидкісні пресети між 1.x і 3.x переписані.

В абсолютних числах: preset 8 у воркері дає **2.4x реального часу** (72 с на 30 с відео).
Година лекції — приблизно 2.4 години на одну сходинку AV1, понад наявну роботу H.264.

Що це змінює: умова (3) стає не питанням пресета, а питанням **образу воркера**. Поки в
ньому SVT-AV1 1.4.1, розмову про AV1 вести нема сенсу незалежно від умов (1) і (2).
Наступний крок, коли до цього дійде, — не «взяти preset 8», а спершу підняти SVT-AV1 у
образі (bookworm-backports, новіша база або власна збірка) і переміряти.

**Обмеження цього виміру.** У воркері **немає `libvmaf`** (лише `vmafmotion`), тож
еквівалентність якості там не перевіряється — виміряно рівно час. І однаковий CRF на
різних мажорних версіях SVT-AV1 не означає однакову якість, тому розміри файлів між
рядками таблиці не порівнюються, лише часи всередині кожного рядка. Хост — macOS, де
`date` не має `%N`; перший прогін дав через це цілі секунди, остаточні хостові числа
переміряні таймером Python.

#### Підняття образу знімає проблему — і має власну ціну

Зібрано пробний образ за тим самим рецептом (`apt-get install ffmpeg`), лише на
`debian:trixie-slim`: ffmpeg 7.1.5, **SVT-AV1 2.3.0**. Продакшн-`Dockerfile` не чіпали —
§24.14 тримає AV1 відкладеним, тож тягнути новішу базу в бойовий образ заради проби зарано.
Backports тут не підходять окремо: ffmpeg у bookworm злінкований із `libsvtav1enc1`, а
новіший пакет має інший soname, тож без перезбирання ffmpeg не підхопиться.

| образ | H.264 | AV1 preset 8 | кратність | реальний час |
|---|---|---|---|---|
| поточний (bookworm, SVT-AV1 1.4.1) | 2.26 с | 54.46 с | **24.1x** | 1.8x |
| пробний (trixie, SVT-AV1 2.3.0) | 2.00 с | 5.47 с | **2.7x** | **0.18x** |

Апгрейд прискорює AV1 вдесятеро й повертає кратність у ту саму зону, що на хості (1.8x при
3.0.2). Година лекції — ~11 хвилин на сходинку AV1 замість ~1.8 години. **Умова (3) стає
досяжною, але лише після підняття образу.**

**Ціна, зміряна тим самим прогоном справжньої драбини (`renderArguments`, 360p+720p).**
Підняття бази — не лише апгрейд кодека: ffmpeg 5.1.9 і 7.1.5 дають **різний master-плейлист**
(це вже фіксував реєстр Phase 5, тепер перевірено виконанням):

| | 5.1.9 | 7.1.5 |
|---|---|---|
| `AVERAGE-BANDWIDTH` | відсутній | **з'явився** |
| `CHANNELS` в `EXT-X-MEDIA` | відсутній | **`CHANNELS="2"`** |
| `BANDWIDTH` 360p / 720p | 542300 / 3440800 | 400416 / 1992900 |

Що з цього ламається, перевірено, а не припущено:

- Регулярка e2e `(?:^|[:,])BANDWIDTH=(\d+)(?:,|$)` **витримує** — дефіс перед
  `AVERAGE-BANDWIDTH` не входить у клас, тож новий атрибут не перехоплюється.
- **R35 зсувається на 35%.** Цей рядок приймання рахує throttle як середнє геометричне
  обох рангів: `sqrt(542300 × 3440800) = 1 365 996` проти `sqrt(400416 × 1992900) = 893 302`
  bps. R35 і так значиться в обмеженнях Phase 6 як крок ABR на синтетичному каналі; зміна
  швидкості каналу на третину може змінити, чи перемикання взагалі спрацює.
- Жодне твердження не спирається на **відсутність** `AVERAGE-BANDWIDTH` чи на вміст
  `EXT-X-MEDIA`, тож поява обох сама по собі нічого не валить. За Apple HLS Authoring Spec
  `AVERAGE-BANDWIDTH` узагалі обов'язковий, тобто вихід 7.1.5 коректніший.

Отже послідовність, коли до AV1 дійде: спершу підняти образ і перепрогнати медійні та
playback-рядки приймання (щонайменше R35), і лише потім говорити про сходинку AV1. Підняття
образу має власну вартість перевірки й не належить до роботи по AV1 — його варто робити
окремим слайсом із власними доказами.

### 6.3 Наскільки менший потік

Виміряно на трьох реальних джерелах (натуральне відео з зерном і рухом; десктопний
скрінкаст; мобільний скрінкаст, HEVC першого покоління), матч по якості через VMAF з
інтерполяцією в точці рівної якості:

| Режим виміру | Економія | Що це означає |
|---|---|---|
| Безкаповий CRF | −20…−57% | «скільки байтів треба AV1 на ту саму якість» — **не** прогноз egress |
| З нашими VBV-капами | **−7…−30%** | те, що реально поїде до клієнта |

Капи з'їдають приблизно половину переваги: capped-CRF у SVT-AV1 розподіляє біти помітно
гірше за чистий CRF (≈1 VMAF гірше за ті самі байти). Для планування брати **нижній край**:
на важкому натуральному відео вийшло лише −7.3%.

Альтернативне прочитання тих самих даних: за **той самий** бюджет AV1 дає +2.4…+11 VMAF.

### 6.4 SVT-AV1 не тримає стелю бітрейту (нове, впливає на реалізацію)

По всіх capped-замірах SVT-AV1 віддавав **96-155%** заданого `maxrate`; `x264` на тому ж
контенті тримався 101%. Для HLS, де cap — гарантія доставки, драбину не можна перенести
на capped-CRF як є: або закладати cap на 25-40% нижче цільового, або брати інший режим
rate control і міряти його окремо. `-svtav1-params mbr=` для цього не годиться — параметр
у kbps, і значення в bps мовчки клампиться до 100 Mbps; працює лише `-maxrate`.

### 6.5 Пастки методології (коштували неправильних чисел, перш ніж їх знайшли)

1. **Keyframe-паритет.** Перша версія форсувала наші 2-секундні keyframes лише на H.264;
   AV1 отримував їх кожні ~10.7 с і виглядав кращим, ніж є. Рахувати keyframes у виході,
   а не припускати.
2. **Модель VMAF розрахована на перегляд 1080p.** Оцінювання 360p/720p у нативній
   роздільності завищує бал і ховає артефакти. Нормалізувати обидва потоки; портретний
   контент — до 1080 **рядків** зі збереженням пропорцій, не розчавлювати в 1920x1080.
3. **Капи для sweep прив'язувати до досягнутого бітрейту H.264, а не до номінального cap
   драбини.** Контент, що їсть 145 kbps із 3000 kbps рангу, під номінальними капами не
   впирається в жоден, і всі рядки виходять однакові.
4. **Синтетична фікстура дає протилежний висновок.** На `e2e-1280x720-100s.mp4` AV1
   preset 8 вийшов **+63% більшим**, і з цього було зроблено висновок, що пресет перевертає
   знак результату. На всіх трьох реальних джерелах preset 8 виграє 25-49%. Не міряти
   кодеки на тест-патернах.

### 6.6 Побічне спостереження: ранг 360p

На натуральному відео з високим рухом H.264 у ранзі 360p упирається в `maxrate 365k`
(369 kbps, 101% cap) і дає **VMAF 60.23** — низько за будь-яким прочитанням. AV1 за той
самий бюджет дає 71.30, тобто пом'якшує, але корінь не в кодеку. Чи це прийнятно за
задумом (найнижча сходинка для поганого каналу), чи ранг недофінансований — відкрите
питання до власника продукту; занесено в `work/backlog.md`.

---

## 5. Джерела

1. [S1] Netflix TechBlog, «AV1 — Now Powering 30% of Netflix Streaming», 2025-12-01 — https://netflixtechblog.com/av1-now-powering-30-of-netflix-streaming-02f592242d80
2. [S2] Netflix TechBlog, «Bringing AV1 Streaming to Netflix Members' TVs», 2021-11-09 — https://netflixtechblog.com/bringing-av1-streaming-to-netflix-members-tvs-b7fc88e42320
3. [S3] Netflix TechBlog, «Netflix Now Streaming AV1 on Android», 2020-02-05 — https://netflixtechblog.com/netflix-now-streaming-av1-on-android-d5264a515202
4. [S4] Netflix TechBlog, «Per-Title Encode Optimization», 2015-12-14 — https://netflixtechblog.com/per-title-encode-optimization-7e99442b62a2
5. [S5] Netflix TechBlog, «Optimized shot-based encodes: Now Streaming!», 2018-03-09 — https://netflixtechblog.com/optimized-shot-based-encodes-now-streaming-4b9464204830
6. [S6] Netflix TechBlog, «HDR10+ Now Streaming on Netflix», 2025-03-24 — https://netflixtechblog.com/hdr10-now-streaming-on-netflix-c9ab1f4bd72b
7. [S7] Apple Developer, «HLS Authoring Specification for Apple Devices» — https://developer.apple.com/documentation/http-live-streaming/hls-authoring-specification-for-apple-devices
8. [S8] Apple Newsroom, «Apple unveils iPhone 15 Pro and iPhone 15 Pro Max», 2023-09-12 — https://www.apple.com/newsroom/2023/09/apple-unveils-iphone-15-pro-and-iphone-15-pro-max/
9. [S9] Apple Newsroom, «Apple unveils M3, M3 Pro, and M3 Max», 2023-10-30 — https://www.apple.com/newsroom/2023/10/apple-unveils-m3-m3-pro-and-m3-max-the-most-advanced-chips-for-a-personal-computer/
10. [S10] WebKit Blog, «WebKit Features in Safari 17.0», 2023-09-18 — https://webkit.org/blog/14445/webkit-features-in-safari-17-0/
11. [S11] Meta / Vodafone / YouTube, «Video Streaming with the AV1 Video Codec in Mobile Devices» (white paper), 2025-09 — https://engineering.fb.com/wp-content/uploads/2025/09/Meta-AV1-White-Paper-FINAL.pdf
12. [S12] YouTube Help, «YouTube recommended upload encoding settings» — https://support.google.com/youtube/answer/1722171?hl=en
13. [S13] YouTube Help, «Choose live encoder settings, bitrates, and resolutions» — https://support.google.com/youtube/answer/2853702?hl=en
14. [S14] Google for Developers, «YouTube Live Streaming Ingestion Protocol Comparison» — https://developers.google.com/youtube/v3/live/guides/ingestion-protocol-comparison
15. [S15] Chrome for Developers, «Media updates in Chrome 70», 2018-09-18 — https://developer.chrome.com/blog/media-updates-in-chrome-70
16. [S16] Android Developers Blog, «The Second Beta of Android 15», 2024-05 — https://android-developers.googleblog.com/2024/05/the-second-beta-of-android-15.html
17. [S17] AOSP, «Android 14 Compatibility Definition» — https://source.android.com/docs/compatibility/14/android-14-cdd
18. [S18] AOMedia, «How Google Uses AV1 to Drive Innovation» — https://aomedia.org/av1-adoption-showcase/google-story/
19. [S19] Engineering at Meta, «How Meta brought AV1 to Reels», 2023-02-21 — https://engineering.fb.com/2023/02/21/video-engineering/av1-codec-facebook-instagram-reels/
20. [S20] Twitch Blog, «Introducing the Enhanced Broadcasting Beta», 2024-01-08 — https://blog.twitch.tv/en/2024/01/08/introducing-the-enhanced-broadcasting-beta
21. [S21] Twitch Blog, «Introducing Dual Format and 2k Streaming on Twitch», 2026-06-17 — https://blog.twitch.tv/en/2026/06/17/introducing-dual-format-and-2k-streaming-on-twitch/
22. [S22] Vimeo Blog, «AV1: Setting a new standard for video codecs», 2019-06-13 — https://vimeo.com/blog/post/av1-new-standard-codecs
23. [S23] Bitmovin, «The 7th Annual Bitmovin Video Developer Report 2023–2024» (PDF) — https://bitmovin.com/downloads/bitmovin-7th-video-developer-report-2023-2024.pdf
24. [S24] Bitmovin Blog, «The State of AV1 Playback Support [2024 Update]», 2024-05-16 — https://bitmovin.com/blog/av1-playback-support/
25. [S25] Bitmovin Blog, «Bitmovin's Guide to Adopting AV1 Encoding», 2023-11-02 — https://bitmovin.com/blog/av1-encoding-guide/
26. [S26] Bitmovin Blog, «Bitmovin Improves Support AV1 Video Encoding for VoD», upd. 2024-02-19 — https://bitmovin.com/blog/bitmovin-improves-av1-video-encoding/
27. [S27] Mux Docs, «Minimize processing time» (standard input) — https://www.mux.com/docs/guides/minimize-processing-time
28. [S28] Mux Changelog, «HEVC Standard Input Now Available to All Users», 2025-10-15 — https://www.mux.com/docs/changelog/hevc-standard-input-general-availability
29. [S29] Mux Docs, «Use video quality levels» — https://www.mux.com/docs/guides/use-video-quality-levels
30. [S30] Mux Docs (повний текст), «Mux Video developer documentation» — https://www.mux.com/docs/guides/video.txt
31. [S31] Amazon Developer, «Device Specifications: Fire TV Streaming Media Player» — https://developer.amazon.com/docs/device-specs/device-specifications-fire-tv-streaming-media-player.html
32. [S32] FlatpanelsHD, «Disney+ restores 4K on select devices by switching codec», 2026-08-04 (вторинне; містить пряму цитату Disney) — https://flatpanelshd.com/news.php?id=1785834485&subaction=showfull
33. [S33] heise online, «Patent dispute: Disney+ restores 4K in Germany», 2026-08-03 (вторинне; містить пряму цитату Disney) — https://www.heise.de/en/news/Patent-dispute-Disney-restores-4K-in-Germany-11395286.html
34. [S34] Firefox Release Notes 67.0, 2019-05-21 — https://www.firefox.com/en-US/firefox/67.0/releasenotes/
35. [S35] FFmpeg Wiki, «Encode/AV1» — https://trac.ffmpeg.org/wiki/Encode/AV1
36. [S36] SVT-AV1 Docs, «Common Questions» — https://gitlab.com/AOMediaCodec/SVT-AV1/-/blob/master/Docs/CommonQuestions.md
37. [S37] Bitmovin Press Room, «Video Developer Report 2025/26», 2025-09-11 — https://bitmovin.com/press-room/bitmovin-com-video-developer-report-2025/

Внутрішні посилання: `docs/media/spec.md` (Q3: H.264/AAC, 360p/720p/1080p, HLS fMP4 `.m4s`),
`docs/architecture/media.md` (згадує MPEG-TS — застаріле відносно Q3), `docs/adr/005-hls-nginx.md`
(обрав TS «for first HLS implementation» — суперечить Q3; ADR потребує оновлення окремим
кроком, ця нотатка його не змінює).
