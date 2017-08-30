/*
Package listReader can parse delimited lists of text encoded floating point numbers.

using the Floats type Read method, parsed values are placed in the provided fixed length buffer, like a Reader a count and error are returned.

there are two classes of delimiters, ones that are single (each is a delimiter) and ones that are multiple (a sequence of any of them is counted as one delimiter).

in this implementation only one, settable, byte is allowed as the 'single' delimiter and the multiples are set to white space, making it appropriate for both human and/or machine readable lists.

common 'single' delimiters might be; ',' ' ' '\t' '\n' '\x1F' '\x00'

parse errors (ParseError type) returned in preference to io.EOF. (io.EOF returned on subsequent Read)

a ParseError does not stop processing, so more than one can occur per Read, the first one is returned.

any items triggering a ParseError are indicated in the returned Float array as the NaN value.

the unconsumed buffer (UnBuf) is exposed.

the delimiter (Delimiter) is exposed.

helper Readers are provided.
*/
package listreader
