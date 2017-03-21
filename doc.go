/*
Package listReader can parse delimited lists of text encoded floating point numbers.

the delimiters and digits interpretations are integrated, and unconsumed bytes are internally buffered, for speed.

results are returned in a fixed length buffer with a count, like a Reader, so files larger than memory can be processed.

it uses two classes of delimiter, ones that are single (each is a delimiter) and ones that are multiple (a sequence of any of them is counted as one delimiter).

in this implementation one, settable, byte is allowed as the 'single' delimiter and the multiples are set to white space, making it appropriate for both human and/or machine readable lists.

common 'single' delimiters might be; ',' ' ' '\t' '\n' '\x1F' '\x00'

an exposed flag indicates if any items failed to parse.

the unconsumed buffer is exposed.

*/
package listreader
