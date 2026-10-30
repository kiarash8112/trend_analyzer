import { Pipe, PipeTransform } from '@angular/core';
import { hostOf } from '../utils/text';

@Pipe({ name: 'host' })
export class HostPipe implements PipeTransform {
  transform(url: string): string {
    return hostOf(url);
  }
}
