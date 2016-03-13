### Website TTFB — Time To First Byte

> Time To First Byte (TTFB) is a measurement used as an indication of the responsiveness of a webserver or other network resource.
>
> TTFB measures the duration from the user or client making an HTTP request to the first byte of the page being received by the client's browser. This time is made up of the socket connection time, the time taken to send the HTTP request, and the time taken to get the first byte of the page. Although sometimes misunderstood as a post-DNS calculation, the original calculation of TTFB in networking always includes network latency in measuring the time it takes for a resource to begin loading.
>
> Often, a smaller (faster) TTFB size is seen as a benchmark of a well-configured server application. For example, a lower Time To First Byte could point to fewer dynamic calculations being performed by the web-server, although this is often due to caching at either the DNS, server, or application level. More commonly, a very low TTFB is observed with statically served web pages, while larger TTFB is often seen with larger, dynamic data requests being pulled from a database.
>
> — More information at [Time To First Byte, Wikipedia](https://en.wikipedia.org/wiki/Time_To_First_Byte)
