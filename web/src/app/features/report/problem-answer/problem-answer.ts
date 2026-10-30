import { Component, input } from '@angular/core';

/** A two-stop line: the problem people have, then this repo as the answer. */
@Component({
  selector: 'app-problem-answer',
  templateUrl: './problem-answer.html',
  styleUrl: './problem-answer.css',
})
export class ProblemAnswer {
  readonly problem = input.required<string>();
  readonly solution = input.required<string>();
  readonly repoName = input.required<string>();
}
