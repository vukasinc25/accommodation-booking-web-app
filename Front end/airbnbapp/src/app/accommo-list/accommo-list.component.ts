import { Component, Input } from '@angular/core';
import { Accommodation } from '../model/accommodation';

@Component({
  selector: 'app-accommo-list',
  templateUrl: './accommo-list.component.html',
  styleUrls: ['./accommo-list.component.css'],
})
export class AccommoListComponent {
  @Input() accommodations: Accommodation[] = [];
}
