# tranco
package to access to the Tranco list, published at https://tranco-list.eu.

# Usage

import github.com/mustafaocak/tranco

Create a variable of type TrancoList
 
var tl tranco.TrancoList

To get the latest tranco list

t.Should_cache = true
tl = t.List("latest")

To get the rank info a domain in the list

fmt.Println(tl.Rank("google.com"))

To get top 10 elements of the list

fmt.Println(tl.Top(10))
