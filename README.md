This repo has three files:

	main.go (solution)

	results.csv (output)

	urls.csv (raw data)
___________________________

File urls.csv has URL

	http://i.imgur.com/enhDnTM.jpg

twenty-five (25) times. The first twelve times the genHash2() function sees this file, it calculates one hash value. For the last thirteen times, genHash2() calculates a different hash value. Because of this, results.csv has

	http://i.imgur.com/enhDnTM.jpg

twice. I did not figure out why genHash2() returned two different hash values.