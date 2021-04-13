# ShakeSearch

Welcome to the Pulley Shakesearch Take-home Challenge! In this repository,
you'll find a simple web app that allows a user to search for a text string in
the complete works of Shakespeare.

You can see a live version of the app at
https://pulley-shakesearch.herokuapp.com/. Try searching for "Hamlet" to display
a set of results.

In it's current state, however, the app is just a rough prototype. The search is
case sensitive, the results are difficult to read, and the search is limited to
exact matches.

## Your Mission

Improve the search backend. Think about the problem from the **user's perspective**
and prioritize your changes according to what you think is most useful. 

## Evaluation

We will be primarily evaluating based on how well the search works for users. A search result with a lot of features (i.e. multi-words and mis-spellings handled), but with results that are hard to read would not be a strong submission. 


## Submission

1. Fork this repository and send us a link to your fork after pushing your changes. 
2. Heroku hosting - The project includes a Heroku Procfile and, in its
current state, can be deployed easily on Heroku's free tier.
3. In your submission, share with us what changes you made and how you would prioritize changes if you had more time.



---
---

## Changes

- Overhauled the way the search worked. My search uses the Levenshtein distance method via the [FuzzySearch](https://github.com/lithammer/fuzzysearch) library to find all matches that are similar to the search query.  
- The search results are displayed sorted by their Levenshtein distance, and in cases where the Levenshtein distance is the same, they are displayed in their order of appearance in the text file. 
- I also changed the way search results are displayed. The search result is now accompanied by the line number where the result starts and the search result includes the entire dialog instead of just 500 characters and the word that gets matched is highlighted.
- The search is done on each word in the query separately as Fuzzy Search doesn't do well when using the entire search query as a whole.
- The app is hosted on heroku at this [link](https://rohitc-shakesearch.herokuapp.com/)

## Future Work

- As of now, the space complexity is quite high as the complete text is being stored in a couple of different formats, if I had more time, I would try to store the text only once in a format usable for in all the ways I want to.
- As a problem, I think the most efficient way to process the results would be to have some substantial amount of pre-processing on the text and I would have opted to pre-process to extract things like Chapter Names, Character Names and such to have a richer results display.
- I would also refactor the code to have proper packages for the smaller utility functions, right now the code looks very messy.
- I have opted for a very simplistic approach to the way results are displayed, I would definitely clean the display page a bit as well.  