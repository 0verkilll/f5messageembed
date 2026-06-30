# f5messageembed

Pure-Go F5 steganography: embed a hidden message into the DCT coefficients of a JPEG. Uses SHA1PRNG, which is deterministic and **not cryptographically secure**.

## Install

```bash
go get github.com/0verkilll/f5messageembed
```

## Command-line tool

Prebuilt binaries for every platform are attached to each [release](https://github.com/0verkilll/f5messageembed/releases).

```
f5embed -in cover.jpg -out stego.jpg -pass <password> -msg "your secret"
```

## Sponsor

If this project is useful to you, please consider supporting its development:

**[Sponsor @0verkilll on GitHub](https://github.com/sponsors/0verkilll)**

## License

[MIT](LICENSE)
