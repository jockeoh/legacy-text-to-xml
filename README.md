# Legacy text to XML converter

A small Go command-line program that converts a line-based people format to XML.

## Run

Requires Go 1.27.1.

```sh
go run ./cmd/converter input.txt output.xml
```

Or build a binary:

```sh
go build -o converter ./cmd/converter
./converter input.txt output.xml
```

## Input format

```text
P|firstname|lastname
T|mobile|landline
A|street|city|zip
F|name|born
```

`P` starts a person and `F` starts a family member. A following phone or address belongs to the current family member; otherwise it belongs to the person. A new `P` clears the family context.

Addresses normally contain a postcode, but the supplied example includes `A|10 Downing Street|London`. The converter therefore accepts addresses with or without zip. A missing zip is omitted from the XML, while an explicitly empty field produces an empty `<zip>` element.

Blank lines are ignored. Unknown records, invalid field counts, invalid hierarchy and duplicate phone/address records cause an error with the source line number. Field values are kept as provided; the converter does not validate phone numbers, postcodes or birth years.

The source format defines no escaping for `|`, so the delimiter cannot occur inside a field.

## Output

```xml
<people>
    <person>
        <firstname>Elof</firstname>
        <lastname>Sundin</lastname>
        <address>
            <street>S:t Johannesgatan 16</street>
            <city>Uppsala</city>
            <zip>75330</zip>
        </address>
        <phone>
            <mobile>073-101801</mobile>
            <landline>018-101801</landline>
        </phone>
    </person>
</people>
```

Addresses are written before phones regardless of input order. Family members keep their input order. XML escaping is handled by `encoding/xml`; no XML declaration is written.

The complete example is available in `testdata/`.

## Design

The parser reads from an `io.Reader` into a small in-memory model. A separate writer serializes the model to an `io.Writer`. File handling stays in the CLI under `cmd/converter`, while the parser, model and XML writer are kept under `internal`.

`bufio.Scanner` is configured with a 1 MiB maximum line size. The current implementation keeps the complete result in memory; if input files became large, the next `P` record could be used as a boundary for writing one person at a time.

## Tests

```sh
go test ./...
go vet ./...
go build ./...
```
