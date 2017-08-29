/*
Package listReader can parse delimited lists of text encoded floating point numbers.

it uses two classes of delimiters, ones that are single (each is a delimiter) and ones that are multiple (a sequence of any of them is counted as one delimiter).

in this implementation only one, settable, byte is allowed as the 'single' delimiter and the multiples are set to white space, making it appropriate for both human and/or machine readable lists.

designed for a list of the same type, not well suited to records two level stucture.

common 'single' delimiters might be; ',' ' ' '\t' '\n' '\x1F' '\x00'

results are returned in the provided fixed length buffer with a count, like a Reader.

an exposed flag indicates if ANY items failed to parse and these are represented in the buffer as NAN values.

the unconsumed buffer is exposed.

*/
package listreader
