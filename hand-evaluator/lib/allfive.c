#include "poker.h"
#include <stdio.h>

/****************************************************************
    This code tests my evaluator by looping over all 2,598,960
    possible five-card poker hands, calculating each hand's
    distinct value, and displaying the frequency count of each
    hand type.  It also prints the amount of time taken to
    perform all the calculations.

    Kevin L. Suffecool (a.k.a "Cactus Kev"), 2001
    kevin@suffe.cool
****************************************************************/

// :3
// -- LainVT
int main(int argc, char **argv) {
  int eval = -1;
  if (argc == 6) {
    int hand[5];
    for (int i = 0; i < 5; i++) {
      hand[i] = get_weird_number(argv[i + 1][0], argv[i + 1][1]);
    }

    eval = eval_5hand(hand);
  } else if (argc == 8) {
    int hand[7];
    for (int i = 0; i < 7; i++) {
      hand[i] = get_weird_number(argv[i + 1][0], argv[i + 1][1]);
    }

    eval = eval_7hand(hand);
  } else {
    fprintf(stderr, "called with %d args, requires 5 or 7\n", argc - 1);
    return 6;
  }

  if (eval == 0) {
    fprintf(stderr, "invalid hand provided\n");
    return 7;
  }

  fprintf(stdout, "%d\n", eval);
  return 0;
}
