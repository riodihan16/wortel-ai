"""Regenerate the app icon (build/appicon.png + build/windows/icon.ico).

Draws the carrot character from frontend/src/components/ui/Carrot.tsx
(viewBox 116, fixed --carrot-* palette) so the exe, the taskbar and the tray
all show the same mascot as the pet window.

Run from the repo root:  python scripts/make-icon.py
Needs Pillow. Sizes under FACE_MIN_PX are drawn faceless: a face is mush at
16-32 px and the silhouette plus leaves is what reads there.
"""
from pathlib import Path

from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parent.parent
SS = 8  # supersample factor per output size
VIEW = 116.0
FACE_MIN_PX = 48
ICO_SIZES = [16, 20, 24, 32, 40, 48, 64, 128, 256]

# sRGB equivalents of the --carrot-* oklch tokens in styles/tokens.css.
BODY = (246, 140, 54)
SHADE = (196, 86, 30)
LEAF = (88, 186, 84)
INK = (58, 32, 22)
CHEEK = (232, 92, 72)
SHINE = (255, 255, 255)


def bezier(p0, p1, p2, p3, n=24):
    pts = []
    for i in range(n + 1):
        t = i / n
        a = (1 - t) ** 3
        b = 3 * (1 - t) ** 2 * t
        c = 3 * (1 - t) * t**2
        d = t**3
        pts.append((a * p0[0] + b * p1[0] + c * p2[0] + d * p3[0],
                    a * p0[1] + b * p1[1] + c * p2[1] + d * p3[1]))
    return pts


def quad(p0, p1, p2, n=16):
    return [((1 - t) ** 2 * p0[0] + 2 * (1 - t) * t * p1[0] + t**2 * p2[0],
             (1 - t) ** 2 * p0[1] + 2 * (1 - t) * t * p1[1] + t**2 * p2[1])
            for t in (i / n for i in range(n + 1))]


def leaf_polys():
    return [
        bezier((58, 34), (48, 28), (42, 18), (42, 8)) + bezier((42, 8), (52, 10), (58, 21), (58, 34)),
        bezier((58, 32), (57, 22), (57, 12), (60, 4)) + bezier((60, 4), (64, 12), (63, 24), (58, 32)),
        bezier((58, 34), (66, 30), (72, 22), (73, 13)) + bezier((73, 13), (64, 14), (58, 24), (58, 34)),
    ]


def body_poly():
    pts = [(38, 28), (78, 28)]
    pts += bezier((78, 28), (78, 50), (68, 74), (59, 99))
    pts += bezier((59, 99), (58.4, 101), (57.6, 101), (57, 99))
    pts += bezier((57, 99), (48, 74), (38, 50), (38, 28))
    return pts


def render(px: int) -> Image.Image:
    big = px * SS
    # Fit the 116-unit drawing (content spans ~x36..80, y3..102) with a margin.
    scale = big * 0.92 / 100.0
    ox = big / 2 - 58 * scale
    oy = big * 0.04 - 3 * scale

    def tf(pts):
        return [(ox + x * scale, oy + y * scale) for x, y in pts]

    im = Image.new("RGBA", (big, big), (0, 0, 0, 0))
    d = ImageDraw.Draw(im)
    stroke = max(2, int(2.6 * scale))

    def poly(pts, fill):
        d.polygon(tf(pts), fill=fill)
        d.line(tf(pts) + [tf(pts)[0]], fill=INK, width=stroke, joint="curve")

    for lp in leaf_polys():
        poly(lp, LEAF)
    poly(body_poly(), BODY)

    # Shine down the left flank, soft.
    shine = Image.new("RGBA", (big, big), (0, 0, 0, 0))
    sd = ImageDraw.Draw(shine)
    sp = [(44, 31), (50, 31)] + bezier((50, 31), (48.5, 50), (47.5, 68), (49.5, 84)) \
        + bezier((49.5, 84), (44.5, 70), (42.5, 48), (44, 31))
    sd.polygon(tf(sp), fill=SHINE + (66,))
    im = Image.alpha_composite(im, shine)
    d = ImageDraw.Draw(im)

    # Grooves.
    gw = max(1, int(2 * scale))
    for g in (quad((63, 68), (67.5, 68), (70, 65.5)), quad((47, 76), (51, 78), (55, 77.5))):
        d.line(tf(g), fill=SHADE, width=gw, joint="curve")

    if px >= FACE_MIN_PX:
        for cx in (51.5, 64.5):
            r = 3.2 * scale
            x, y = ox + cx * scale, oy + 46 * scale
            d.ellipse([x - r, y - r, x + r, y + r], fill=INK)
        for cx in (47.5, 68.5):
            r = 3.4 * scale
            x, y = ox + cx * scale, oy + 54 * scale
            d.ellipse([x - r, y - r, x + r, y + r], fill=CHEEK + (150,))
        d.line(tf(quad((52, 56), (58, 61.5), (64, 56))), fill=INK, width=max(2, int(2 * scale)), joint="curve")

    return im.resize((px, px), Image.LANCZOS)


def main():
    (ROOT / "build").mkdir(exist_ok=True)
    render(1024).save(ROOT / "build" / "appicon.png")
    imgs = [render(s) for s in ICO_SIZES]
    imgs[-1].save(
        ROOT / "build" / "windows" / "icon.ico",
        format="ICO",
        sizes=[(s, s) for s in ICO_SIZES],
        append_images=imgs[:-1],
    )
    import os

    out = os.environ.get("ICON_PREVIEW")
    if out:  # preview sheet for eyeballing; not used by the build
        sheet = Image.new("RGBA", (256 + 20 + 128 + 20 + 48 + 10 + 32 + 10 + 16 + 20, 276), (60, 60, 64, 255))
        x = 10
        for s_ in (256, 128, 48, 32, 16):
            sheet.alpha_composite(render(s_), (x, 10))
            x += s_ + 10
        sheet.save(out)


if __name__ == "__main__":
    main()
