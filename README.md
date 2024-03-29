# versioncmp

A version string comparison library for golang.

While there are libraries for semver, I couldn't find anything that handles
"everything".

The reason I wanted this, is another project of mine, called
[spoon](https://github.com/Bios-Marcel/spoon). One of its tasks, is to parse
sscoop repository manifest versions, which come from a ton of different
developers, each of them bringing their own unqie versioning concept. Isn't that
great?

## It can't compare version format X

You've got yet another great format? We can try incorporating it. As long as we
don't break anything else!

Feel free to contribute your own. However, lack of new tests means I'll close
your PR without a comment.

## Usage

First, grab the library:

```
go get github.com/Bios-Marcel/versioncmp
```

There's currently only a single public function:

```go
fmt.Println(versioncmp.Compare("1.0.0", "2.0.0")) //2.0.0
fmt.Println(versioncmp.Compare("1.0.0", "1.0.0")) //""
```

## How it works

It returns the `greater` of the two versions.

The concept it uses is to group parts of the version together, depending on
which separator was last met. For example `1.0.0-2.5` would lead to the
grouping `[[1,0,0],[2,5]]`. These groups are the compared with the other
versions groups. The first group that is bigger, wins.

In addition to the groups, we try to find out whether there's something I call a
`stability` (there's probably a better word, but I don't care). The stabilities
are:
  * stable
  * pre
  * rc
  * beta
  * alpha
  * dev

These are sorted from `stable` to `unstable`. If the number groups are equal,
the stability is used as a second factor.

## Options

Currently there are two options. `CompareNightly` and `CompareMeta`. Generally
nightly versions are always deemed equal. Metadata is always ignored. Even if
you wish to compare it, we only have `equal` or `unequal`, instead of `greater`,
`equal` and `less`.

