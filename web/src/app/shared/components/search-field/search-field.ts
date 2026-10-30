import { Component, input, model } from '@angular/core';

@Component({
  selector: 'app-search-field',
  templateUrl: './search-field.html',
  styleUrl: './search-field.css',
})
export class SearchField {
  readonly value = model('');
  readonly placeholder = input('');
  readonly label = input.required<string>();
}
