import base64
import json
import math
import sys
from io import BytesIO

import imageio.v3 as imageio
import numpy as np
from PIL import Image, ImageDraw


def fail(message):
    json.dump({"error": {"message": message}}, sys.stdout, separators=(",", ":"))
    sys.stdout.write("\n")
    sys.exit(1)


def require_int(args, name, default, minimum, maximum):
    value = args.get(name, default)
    if not isinstance(value, int):
        fail(f"{name} must be an integer")
    if value < minimum or value > maximum:
        fail(f"{name} must be between {minimum} and {maximum}")
    return value


def require_text(args, name, default):
    value = args.get(name, default)
    if not isinstance(value, str):
        fail(f"{name} must be a string")
    return value


def frame_color(t):
    red = int(70 + 120 * (0.5 + 0.5 * math.sin(t * 2 * math.pi)))
    green = int(110 + 90 * (0.5 + 0.5 * math.sin(t * 2 * math.pi + 2.0)))
    blue = int(170 + 70 * (0.5 + 0.5 * math.sin(t * 2 * math.pi + 4.0)))
    return red, green, blue


def draw_star(draw, center, radius, fill, outline):
    points = []
    for index in range(10):
        angle = -math.pi / 2 + index * math.pi / 5
        current = radius if index % 2 == 0 else radius * 0.45
        points.append(
            (
                center[0] + math.cos(angle) * current,
                center[1] + math.sin(angle) * current,
            )
        )
    draw.polygon(points, fill=fill, outline=outline)


def build_gif(args):
    width = require_int(args, "width", 128, 64, 480)
    height = require_int(args, "height", 128, 64, 480)
    fps = require_int(args, "fps", 12, 1, 30)
    frames = require_int(args, "frames", 18, 2, 90)
    colors = require_int(args, "colors", 64, 8, 256)
    label = require_text(args, "label", "")
    theme = require_text(args, "theme", "star")

    rendered = []
    for index in range(frames):
        t = index / frames
        background = frame_color(t)
        image = Image.new("RGB", (width, height), background)
        draw = ImageDraw.Draw(image)

        pulse = 0.78 + 0.22 * math.sin(t * 2 * math.pi)
        radius = min(width, height) * (0.26 + 0.08 * pulse)
        center = (width / 2, height / 2)
        glow = int(radius * 1.35)
        draw.ellipse(
            [center[0] - glow, center[1] - glow, center[0] + glow, center[1] + glow],
            fill=(255, 255, 255),
        )

        if theme == "dot":
            draw.ellipse(
                [center[0] - radius, center[1] - radius, center[0] + radius, center[1] + radius],
                fill=(255, 210, 64),
                outline=(40, 40, 40),
                width=max(2, int(width / 64)),
            )
        else:
            draw_star(draw, center, radius, fill=(255, 220, 70), outline=(45, 45, 45))

        if label:
            text = label[:12]
            box = draw.textbbox((0, 0), text)
            text_width = box[2] - box[0]
            draw.text(((width - text_width) / 2, height * 0.76), text, fill=(20, 20, 20))

        rendered.append(np.array(image.quantize(colors=colors).convert("RGB")))

    out = BytesIO()
    imageio.imwrite(out, rendered, extension=".gif", duration=1000 / fps, loop=0)
    body = out.getvalue()
    return {
        "mime_type": "image/gif",
        "encoding": "base64",
        "data": base64.b64encode(body).decode("ascii"),
        "size_bytes": len(body),
        "width": width,
        "height": height,
        "frames": frames,
        "fps": fps,
    }


def main():
    raw = sys.stdin.buffer.read()
    try:
        args = json.loads(raw.decode("utf-8")) if raw else {}
    except json.JSONDecodeError as exc:
        fail(f"arguments must be valid JSON: {exc}")
    if not isinstance(args, dict):
        fail("arguments must be a JSON object")
    result = build_gif(args)
    json.dump(result, sys.stdout, ensure_ascii=False, separators=(",", ":"))
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
