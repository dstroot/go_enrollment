# Enrollment Record Processing

### This repository contains code to process enrollment records.

Processing outline:

* Files are sent to us via SFTP.
        - Files can either be in flat or XML format
* Read in and parse the file
        - Validate the file structure - does header match?
* Iterate through the enrollment records
        - Validate the incoming data
                - Valid format
                - Valid Values
                - Etc.
* Insert the data into SQL DB
        - Is this EFIN known to us? If so, update it with new information, otherwise create it. (Should we just load the entire FOIA list?)
                - Does the office name match? (fuzzy)
                - Does the address match? (fuzzy)
                - Etc.
        - Insert a tax year specific record for this enrollment year to indicate the ERO has enrolled for this tax year.
        - Is the EFIN owner known to us and are they associated with this EFIN? IF so just update their information. Otherwise insert the EFIN Owner record




Table of Contents
-----------------

- [Technology](#technology)
- [Performance](#performance)
- [Security](#security)
- [Getting Started](#getting-started)
- [Making Changes](#making-changes)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [TODO](#todo)
- [References](#references)

Technology
----------

We use a **Modern Technology Stack**

+ [Gulp.js](http://gulpjs.com/) Automated build system
+ [Jade](http://jade-lang.com/) Jade is a high performance template engine heavily influenced by Haml and implemented with JavaScript for node and browsers. (all nicely laid out: head, navigation, footer, etc.)
+ [{less}](http://lesscss.org/) Less is a CSS pre-processor, meaning that it extends the CSS language, adding features that allow variables, mixins, functions and many other techniques that allow you to make CSS that is more maintainable, themable and extendable.
+ [Bootstrap](http://getbootstrap.com/) Bootstrap is the most popular HTML, CSS, and JS framework for developing responsive, mobile first projects on the web.
+ [FontAwesome](https://fortawesome.github.io/Font-Awesome/) Font Awesome gives you scalable vector icons that can instantly be customized — size, color, drop shadow, and


Performance
-----------


Security
--------


Getting Started
---------------

Dependencies:
* Go needs to be installed
* Git needs to be installed

Steps to install and run:

```
# Clone the repo (and fetch only the latest commits)
git clone --depth=1 https://github.com/sbtpg/tpg-landing.git
cd tpg-landing

# Install dependencies
npm install

# Run it!
gulp
```

Making Changes
--------------

This is the basic outline to make changes to our codebase.  

1. There are two main branches of code, "Master" and "Development".  Master is what is deployed into production, Development is for development and QA.  To make changes make sure you are on the development branch. `$ git checkout development`
2. Always pull down from GitHub any changes others may have made first: `$ git pull`
3. **When you have stable working code and have unit tested your changes** then check your changes with `$ git status` to double check what files have changed.  Then add the changes: `$ git add -A` and commit the changes with a descriptive message (see references below): `$ git commit -m "Descriptive Commit Message"`
4. Push the new changes up to GitHub: `$ git push`
5. Create a pull request using GitHub (desktop software or the website) and describe the changes you have made, referencing any open issues addressed by your changes.
6. QA will be notified there is a pull request via GitHub, they will pull down the changes and test them. `$ git checkout development`, `$ git pull` and `$ gulp`
7. If testing is successful QA will merge the pull request using GitHub.
8. Then QA will pull down master and push it to production: `$ git checkout master`, `$ git pull` and `$ gulp publish`

:exclamation: **Note:**

* Pull requests do not have a 1:1 relationship with issues.  
* Pull requests can be for code refactoring, etc. so not every pull request will correspond to open issue(s).
* On the other hand, several issues can be closed with one pull request.  Please make sure they are referenced in your pull request so they can be closed.  

Deployment
----------

To upload changes to Amazon AWS S3 run `gulp publish` or `npm publish`.  If you do this in the development branch it will deploy into QA. If you do this in the master branch it will deploy to production.

:exclamation: **Note**: Copy `.env.example` to `.env` and fill in your TPG AWS credentials first for this to work.

Contributing
------------

All you have to do is follow the steps in "making changes".  The code linters and quality tools will catch errors and enforce coding style.

TODO
----

+ Automate deployment from merge into master branch?
+ Control who is allowed to merge?

References
----------

* [Writing good commit messages](https://github.com/erlang/otp/wiki/Writing-good-commit-messages)
* [How to Write a Git Commit Message](http://chris.beams.io/posts/git-commit/)
